// internal/transport/http/api/trigger_api.go（或新文件 sc_trigger_api.go）
package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/contracts"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type ScTriggerAPI struct {
	ch        chan contracts.ScTriggerMsg
	scSvc     *sc.ScService
	scRootDir string
}

func NewScTriggerAPI(deps *svc.Deps) *ScTriggerAPI {
	return &ScTriggerAPI{
		ch:        deps.ScTrigger,
		scSvc:     sc.NewScService(deps),
		scRootDir: deps.Config.Film.ScRootDir,
	}
}

// POST /api/triggers/sc/move { "scName": "xxx" }
type scMoveReq struct {
	ScName string `json:"scName" form:"scName"`
}

func (h *ScTriggerAPI) Move(c *gin.Context) {
	var req scMoveReq
	_ = c.ShouldBind(&req)
	name := strings.TrimSpace(req.ScName)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scName required"})
		return
	}
	msg := contracts.ScTriggerMsg{Kind: contracts.ScMove, ScName: name}
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "sc trigger queue is full"})
	}
}

func (h *ScTriggerAPI) Add(c *gin.Context) {
	var req scAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	dir, err := resolveSingleScEventDir(h.scRootDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg := contracts.ScTriggerMsg{
		Kind:           contracts.ScAdd,
		Dir:            dir,
		ComeMovieJavId: strings.TrimSpace(req.ComeMovieJavId),
		MovieCast:      strings.TrimSpace(req.MovieCast),
		Duration:       req.Duration,
		Fg:             strings.TrimSpace(req.Fg),
		Vessel:         strings.TrimSpace(req.Vessel),
		Remarks:        strings.TrimSpace(req.Remarks),
	}
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "sc trigger queue is full"})
	}
}

type scAddReq struct {
	ComeMovieJavId string `json:"comeMovieJavId"`
	MovieCast      string `json:"movieCast"`
	Duration       int64  `json:"duration"`
	Fg             string `json:"fg"`
	Vessel         string `json:"vessel"`
	Remarks        string `json:"remarks"`
}

type scAddPreviewMovieResp struct {
	MovieName  string   `json:"movieName"`
	MovieJavId string   `json:"movieJavId"`
	Casts      []string `json:"casts"`
}

type scAddPreviewResp struct {
	Dir        string                  `json:"dir"`
	ScName     string                  `json:"scName"`
	ScTime     int64                   `json:"scTime"`
	MovieCount int64                   `json:"movieCount"`
	ImageFound bool                    `json:"imageFound"`
	ImageName  string                  `json:"imageName"`
	Movies     []scAddPreviewMovieResp `json:"movies"`
}

func (h *ScTriggerAPI) AddPreview(c *gin.Context) {
	dir, err := resolveSingleScEventDir(h.scRootDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preview, err := h.scSvc.BuildAddScPreview(c.Request.Context(), dir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := scAddPreviewResp{
		Dir:        preview.Dir,
		ScName:     preview.ScName,
		ScTime:     preview.ScTime,
		MovieCount: preview.MovieCount,
		ImageFound: preview.ImageFound,
		ImageName:  preview.ImageName,
		Movies:     make([]scAddPreviewMovieResp, 0, len(preview.Movies)),
	}
	for _, movie := range preview.Movies {
		if movie == nil {
			continue
		}
		resp.Movies = append(resp.Movies, scAddPreviewMovieResp{
			MovieName:  movie.MovieName,
			MovieJavId: movie.MovieJavId,
			Casts:      movie.Casts,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func resolveSingleScEventDir(root string) (string, error) {
	dir := filepath.Clean(strings.TrimSpace(root))
	if dir == "" {
		return "", fmt.Errorf("film.scRootDir is empty")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read sc root dir failed: %w", err)
	}

	valid := make([]string, 0, 4)
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !isValidScEventDirName(name) {
			continue
		}
		valid = append(valid, name)
	}
	sort.Strings(valid)

	if len(valid) == 0 {
		return "", fmt.Errorf("no valid sc event dir found under %s", dir)
	}
	if len(valid) > 1 {
		return "", fmt.Errorf("multiple valid sc event dirs found under %s: %s", dir, strings.Join(valid, ", "))
	}

	fullDir := filepath.Clean(filepath.Join(dir, valid[0]))
	rel, err := filepath.Rel(dir, fullDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("resolved sc dir escapes root")
	}
	info, err := os.Stat(fullDir)
	if err != nil {
		return "", fmt.Errorf("stat sc event dir failed: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("resolved sc event is not a directory")
	}
	return fullDir, nil
}

func isValidScEventDirName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	_, err := time.Parse("2006-01-02-15-04", name)
	return err == nil
}

func (h *ScTriggerAPI) RebuildStats(c *gin.Context) {
	msg := contracts.ScTriggerMsg{Kind: contracts.ScRebuildStats}
	select {
	case h.ch <- msg:
		c.Status(http.StatusAccepted)
	default:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "sc trigger queue is full"})
	}
}

type scPickReq struct {
	Weight int64                      `json:"weight"`
	Req    types.ListMovieFullRequest `json:"req"`
}

type scPickCopyReq struct {
	PickN int         `json:"pickN"`
	Reqs  []scPickReq `json:"reqs"`
}

type scSmartPickReq struct {
	PickN   int                 `json:"pickN"`
	Options sc.SmartPickOptions `json:"options"`
	Reqs    []scPickReq         `json:"reqs"`
}

type scPickCopyCast struct {
	Name     string `json:"name"`
	NameShow string `json:"name_show"`
}

type scPickCopyMovie struct {
	Name                 string           `json:"name"`
	Title                string           `json:"title"`
	ReleasingDate        string           `json:"releasing_date"`
	FilmBirthDate        string           `json:"film_birth_date"`
	Score                float64          `json:"score"`
	ViewersNumberWatched int64            `json:"viewers_number_watched"`
	Director             string           `json:"director"`
	Genre                []string         `json:"genre"`
	Cast                 []scPickCopyCast `json:"cast"`
	JavUrl               string           `json:"jav_url"`
	VideoUrl             string           `json:"video_url"`
	SearchUrl            string           `json:"search_url"`
	BusUrl               string           `json:"bus_url"`
	JacketImg            string           `json:"jacket_img"`
	Owned                int64            `json:"owned"`
	Prefix               string           `json:"prefix"`
	ScTimes              int64            `json:"sc_times"`
	ComeTimes            int64            `json:"come_times"`
	HighestRank          int64            `json:"highest_rank"`
}

func buildPickCopyMovies(movies []*types.MovieType) []scPickCopyMovie {
	if len(movies) == 0 {
		return nil
	}
	resp := make([]scPickCopyMovie, 0, len(movies))
	for _, m := range movies {
		if m == nil {
			continue
		}
		casts := make([]scPickCopyCast, 0, len(m.Cast))
		for _, c := range m.Cast {
			if c == nil {
				continue
			}
			casts = append(casts, scPickCopyCast{
				Name:     c.Name,
				NameShow: c.NameShow,
			})
		}
		resp = append(resp, scPickCopyMovie{
			Name:                 m.Name,
			Title:                m.Title,
			ReleasingDate:        m.ReleasingDate,
			FilmBirthDate:        m.FilmBirthDate,
			Score:                m.Score,
			ViewersNumberWatched: m.ViewersNumberWatched,
			Director:             m.Director,
			Genre:                m.Genre,
			Cast:                 casts,
			JavUrl:               m.JavUrl,
			VideoUrl:             m.VideoUrl,
			SearchUrl:            m.SearchUrl,
			BusUrl:               m.BusUrl,
			JacketImg:            m.JacketImg,
			Owned:                m.Owned,
			Prefix:               m.Prefix,
			ScTimes:              m.ScTimes,
			ComeTimes:            m.ComeTimes,
			HighestRank:          m.HighestRank,
		})
	}
	return resp
}

func buildPickedTotalSizeGB(movies []*types.MovieType) string {
	var totalBytes int64
	for _, m := range movies {
		if m == nil || m.VFilm == nil {
			continue
		}
		totalBytes += m.VFilm.Size
	}
	return fmt.Sprintf("%.2f", float64(totalBytes)/(1024*1024*1024))
}

// POST /api/triggers/sc/pick-copy
func (h *ScTriggerAPI) PickCopy(c *gin.Context) {
	var req scPickCopyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.Reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reqs is empty"})
		return
	}

	converted := make([]sc.PickRequestWithWeight, 0, len(req.Reqs))
	for _, r := range req.Reqs {
		converted = append(converted, sc.PickRequestWithWeight{
			Req:    r.Req,
			Weight: r.Weight,
		})
	}

	movies, err := h.scSvc.PickCopyFromRequests(c.Request.Context(), converted, req.PickN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	started, status := h.scSvc.StartCopyAsync(movies)
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(movies),
		"movies":        buildPickCopyMovies(movies),
		"total_size_gb": buildPickedTotalSizeGB(movies),
		"copy_started":  started,
		"copy_status":   status,
	})
}

// POST /api/triggers/sc/pick-only
func (h *ScTriggerAPI) PickOnly(c *gin.Context) {
	var req scPickCopyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.Reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reqs is empty"})
		return
	}

	converted := make([]sc.PickRequestWithWeight, 0, len(req.Reqs))
	for _, r := range req.Reqs {
		converted = append(converted, sc.PickRequestWithWeight{
			Req:    r.Req,
			Weight: r.Weight,
		})
	}

	movies, err := h.scSvc.PickFromRequests(c.Request.Context(), converted, req.PickN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(movies),
		"movies":        buildPickCopyMovies(movies),
		"total_size_gb": buildPickedTotalSizeGB(movies),
	})
}

// POST /api/triggers/sc/pick-smart-only
func (h *ScTriggerAPI) PickSmartOnly(c *gin.Context) {
	var req scSmartPickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.Reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reqs is empty"})
		return
	}

	converted := make([]sc.PickRequestWithWeight, 0, len(req.Reqs))
	for _, r := range req.Reqs {
		converted = append(converted, sc.PickRequestWithWeight{
			Req:    r.Req,
			Weight: r.Weight,
		})
	}

	movies, err := h.scSvc.SmartPickFromRequests(c.Request.Context(), converted, req.PickN, req.Options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(movies),
		"movies":        buildPickCopyMovies(movies),
		"total_size_gb": buildPickedTotalSizeGB(movies),
	})
}

// POST /api/triggers/sc/pick-smart-copy
func (h *ScTriggerAPI) PickSmartCopy(c *gin.Context) {
	var req scSmartPickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.Reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reqs is empty"})
		return
	}

	converted := make([]sc.PickRequestWithWeight, 0, len(req.Reqs))
	for _, r := range req.Reqs {
		converted = append(converted, sc.PickRequestWithWeight{
			Req:    r.Req,
			Weight: r.Weight,
		})
	}

	movies, err := h.scSvc.SmartPickCopyFromRequests(c.Request.Context(), converted, req.PickN, req.Options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	started, status := h.scSvc.StartCopyAsync(movies)
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(movies),
		"movies":        buildPickCopyMovies(movies),
		"total_size_gb": buildPickedTotalSizeGB(movies),
		"copy_started":  started,
		"copy_status":   status,
	})
}

// GET /api/triggers/sc/copy-status
func (h *ScTriggerAPI) CopyStatus(c *gin.Context) {
	status := h.scSvc.CopyStatus()
	c.JSON(http.StatusOK, status)
}

// POST /api/triggers/sc/copy-stop
func (h *ScTriggerAPI) CopyStop(c *gin.Context) {
	stopped := h.scSvc.StopCopy()
	status := h.scSvc.CopyStatus()
	c.JSON(http.StatusOK, gin.H{
		"stopped": stopped,
		"status":  status,
	})
}
