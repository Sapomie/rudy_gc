package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type ScTriggerAPI struct {
	runtime   *loop.FetchLoopService
	scSvc     *sc.ScService
	scRootDir string
}

func NewScTriggerAPI(deps *svc.Deps) *ScTriggerAPI {
	return &ScTriggerAPI{
		runtime:   newCrawlerRuntime(deps),
		scSvc:     sc.NewService(deps),
		scRootDir: deps.Config.Film.ScRootDir,
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
	h.startTask(c, loop.StartTaskRequest{
		TaskType:        loop.TaskScAdd,
		Dir:             dir,
		ComeMovieJavID:  strings.TrimSpace(req.ComeMovieJavId),
		MovieCast:       strings.TrimSpace(req.MovieCast),
		Kind:            strings.TrimSpace(req.Kind),
		DurationMinutes: req.DurationMinutes,
		Fg:              strings.TrimSpace(req.Fg),
		Vessel:          strings.TrimSpace(req.Vessel),
		Remarks:         strings.TrimSpace(req.Remarks),
		Movies:          buildScAddInputMovies(req.Movies),
	})
}

type scAddReq struct {
	ComeMovieJavId  string          `json:"comeMovieJavId"`
	MovieCast       string          `json:"movieCast"`
	Kind            string          `json:"kind"`
	DurationMinutes int64           `json:"duration"`
	Fg              string          `json:"fg"`
	Vessel          string          `json:"vessel"`
	Remarks         string          `json:"remarks"`
	Movies          []scAddMovieReq `json:"movies"`
}

type scAddMovieReq struct {
	MovieJavId string `json:"movieJavId"`
	IsSc       int64  `json:"isSc"`
}

func buildScAddInputMovies(reqs []scAddMovieReq) []sc.AddScInputMovie {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]sc.AddScInputMovie, 0, len(reqs))
	for _, req := range reqs {
		movieJavId := strings.TrimSpace(req.MovieJavId)
		if movieJavId == "" {
			continue
		}
		out = append(out, sc.AddScInputMovie{
			MovieJavId: movieJavId,
			IsSc:       req.IsSc,
		})
	}
	return out
}

type scAddPreviewMovieResp struct {
	MovieName  string   `json:"movieName"`
	MovieJavId string   `json:"movieJavId"`
	MovieHref  string   `json:"movieHref"`
	JacketImg  string   `json:"jacketImg"`
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
			MovieHref:  movie.MovieHref,
			JacketImg:  movie.JacketImg,
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
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskScRebuildStats})
}

func (h *ScTriggerAPI) startTask(c *gin.Context, req loop.StartTaskRequest) {
	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobStarted(c, jobID, req.TaskType)
}

type scPickReq struct {
	Weight int64                      `json:"weight"`
	Req    types.ListMovieFullRequest `json:"req"`
}

type scSmartPickReq struct {
	PickN   int                 `json:"pickN"`
	Source  string              `json:"source"`
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

func buildPickCopyMovies(movies []*types.MovieType, source string) []scPickCopyMovie {
	if len(movies) == 0 {
		return nil
	}
	source = sc.NormalizeSmartPickSource(source)
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
			FilmBirthDate:        sc.SmartPickMovieBirthDate(m, source),
			Score:                m.Score,
			ViewersNumberWatched: m.ViewersNumberWatched,
			Director:             m.Director,
			Genre:                m.Genre,
			Cast:                 casts,
			JavUrl:               m.JavUrl,
			VideoUrl:             sc.SmartPickMovieVideoURL(m, source),
			SearchUrl:            m.SearchUrl,
			BusUrl:               m.BusUrl,
			JacketImg:            m.JacketImg,
			Owned:                sc.SmartPickMovieOwned(m, source),
			Prefix:               m.Prefix,
			ScTimes:              m.ScTimes,
			ComeTimes:            m.ComeTimes,
			HighestRank:          m.HighestRank,
		})
	}
	return resp
}

func buildPickedTotalSizeGB(movies []*types.MovieType, source string) string {
	var totalBytes int64
	for _, m := range movies {
		totalBytes += sc.SmartPickMovieSize(m, source)
	}
	return fmt.Sprintf("%.2f", float64(totalBytes)/(1024*1024*1024))
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

	source := sc.NormalizeSmartPickSource(req.Source)
	result, err := h.scSvc.SmartPickWithInfoFromRequests(c.Request.Context(), converted, req.PickN, req.Options, source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	totalSizeGB := buildPickedTotalSizeGB(result.Movies, source)
	pickInfo := result.Info
	pickInfo.TotalSizeGB = totalSizeGB
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(result.Movies),
		"movies":        buildPickCopyMovies(result.Movies, source),
		"total_size_gb": totalSizeGB,
		"source":        source,
		"pick_info":     pickInfo,
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

	source := sc.NormalizeSmartPickSource(req.Source)
	result, err := h.scSvc.SmartPickWithInfoFromRequests(c.Request.Context(), converted, req.PickN, req.Options, source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	totalSizeGB := buildPickedTotalSizeGB(result.Movies, source)
	pickInfo := result.Info
	pickInfo.TotalSizeGB = totalSizeGB
	started, status := h.scSvc.StartCopyAsync(result.Movies, source)
	c.JSON(http.StatusOK, gin.H{
		"picked":        len(result.Movies),
		"movies":        buildPickCopyMovies(result.Movies, source),
		"total_size_gb": totalSizeGB,
		"source":        source,
		"pick_info":     pickInfo,
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
