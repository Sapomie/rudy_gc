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

	"rudy_gc/internal/dep"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/loop"
	mediaservice "rudy_gc/internal/service/media"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/types"
)

type ScTriggerAPI struct {
	runtime   *loop.FetchLoopService
	scSvc     *sc.ScService
	mediaSvc  *mediaservice.Service
	scRootDir string
}

func NewScTriggerAPI(deps *dep.Dep) *ScTriggerAPI {
	return &ScTriggerAPI{
		runtime:   newCrawlerRuntime(deps),
		scSvc:     sc.NewService(deps),
		mediaSvc:  mediaservice.NewService(deps),
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
		ComeMovieType:   strings.TrimSpace(req.ComeMovieType),
		ComeMovieJavID:  strings.TrimSpace(req.ComeMovieJavId),
		ComeFC2MovieID:  req.ComeFC2MovieId,
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
	ComeMovieType   string          `json:"comeMovieType"`
	ComeMovieJavId  string          `json:"comeMovieJavId"`
	ComeFC2MovieId  int64           `json:"comeFC2MovieId"`
	MovieCast       string          `json:"movieCast"`
	Kind            string          `json:"kind"`
	DurationMinutes int64           `json:"duration"`
	Fg              string          `json:"fg"`
	Vessel          string          `json:"vessel"`
	Remarks         string          `json:"remarks"`
	Movies          []scAddMovieReq `json:"movies"`
}

type scAddMovieReq struct {
	MovieType  string `json:"movieType"`
	MovieJavId string `json:"movieJavId"`
	FC2MovieId int64  `json:"fc2MovieId"`
	IsSc       int64  `json:"isSc"`
}

func buildScAddInputMovies(reqs []scAddMovieReq) []sc.AddScInputMovie {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]sc.AddScInputMovie, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, sc.AddScInputMovie{
			MovieType:  strings.TrimSpace(req.MovieType),
			MovieJavId: strings.TrimSpace(req.MovieJavId),
			FC2MovieId: req.FC2MovieId,
			IsSc:       req.IsSc,
		})
	}
	return out
}

type scAddPreviewMovieResp struct {
	MovieType  string   `json:"movieType"`
	MovieName  string   `json:"movieName"`
	MovieJavId string   `json:"movieJavId"`
	FC2MovieId int64    `json:"fc2MovieId"`
	MovieKey   string   `json:"movieKey"`
	Title      string   `json:"title"`
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
			MovieType:  movie.MovieType,
			MovieName:  movie.MovieName,
			MovieJavId: movie.MovieJavId,
			FC2MovieId: movie.FC2MovieId,
			MovieKey:   movie.MovieKey,
			Title:      movie.Title,
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

type actorNameReq struct {
	ActorName string `json:"actorName"`
}

func (h *ScTriggerAPI) RebuildActorRank(c *gin.Context) {
	var req actorNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	actorName := strings.TrimSpace(req.ActorName)
	if actorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "actorName is required"})
		return
	}

	h.startTask(c, loop.StartTaskRequest{
		TaskType:  loop.TaskSpiderRebuildActorRank,
		ActorName: actorName,
	})
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
	FC2Req sc.FC2SmartPickRequest     `json:"fc2Req"`
}

type scSmartPickJAVReq struct {
	PickN   int                 `json:"pickN"`
	Options sc.SmartPickOptions `json:"options"`
	Reqs    []scPickReq         `json:"reqs"`
}

type scSmartPickFC2Req struct {
	PickN   int                    `json:"pickN"`
	Options sc.FC2SmartPickOptions `json:"options"`
	Reqs    []scPickReq            `json:"reqs"`
}

type scSmartPickReq struct {
	Source string            `json:"source"`
	JAV    scSmartPickJAVReq `json:"jav"`
	FC2    scSmartPickFC2Req `json:"fc2"`
}

type scPickCopyCast struct {
	PersonID int64  `json:"person_id"`
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
				PersonID: c.PersonId,
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

func buildPickedTotalSize(movies []*types.MovieType, source string) int64 {
	var totalBytes int64
	for _, m := range movies {
		totalBytes += sc.SmartPickMovieSize(m, source)
	}
	return totalBytes
}

func formatSmartPickSizeGB(totalBytes int64) string {
	return fmt.Sprintf("%.2f", float64(totalBytes)/(1024*1024*1024))
}

func (h *ScTriggerAPI) PickSmartOnly(c *gin.Context) {
	var req scSmartPickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.pickSmartMixed(c, req, false)
}

func (h *ScTriggerAPI) PickSmartCopy(c *gin.Context) {
	var req scSmartPickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	h.pickSmartMixed(c, req, true)
}

func validateScSmartPickReq(req scSmartPickReq) error {
	if req.JAV.PickN < 0 {
		return fmt.Errorf("JAV pickN must be >= 0")
	}
	if req.FC2.PickN < 0 {
		return fmt.Errorf("FC2 pickN must be >= 0")
	}
	if req.JAV.PickN == 0 && req.FC2.PickN == 0 {
		return fmt.Errorf("JAV pickN and FC2 pickN cannot both be 0")
	}
	if req.JAV.PickN > 0 && len(req.JAV.Reqs) == 0 {
		return fmt.Errorf("JAV reqs is empty")
	}
	if req.FC2.PickN > 0 && len(req.FC2.Reqs) == 0 {
		return fmt.Errorf("FC2 reqs is empty")
	}
	return nil
}

func (h *ScTriggerAPI) pickSmartMixed(c *gin.Context, req scSmartPickReq, withCopy bool) {
	if err := validateScSmartPickReq(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source := sc.NormalizeSmartPickSource(req.Source)
	var javResult *sc.SmartPickResult
	if req.JAV.PickN > 0 {
		converted := make([]sc.PickRequestWithWeight, 0, len(req.JAV.Reqs))
		for _, item := range req.JAV.Reqs {
			converted = append(converted, sc.PickRequestWithWeight{Req: item.Req, Weight: item.Weight})
		}
		result, err := h.scSvc.SmartPickWithInfoFromRequests(c.Request.Context(), converted, req.JAV.PickN, req.JAV.Options, source)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JAV: " + err.Error()})
			return
		}
		javResult = result
	}

	var fc2Result *sc.FC2SmartPickResult
	if req.FC2.PickN > 0 {
		converted := make([]sc.FC2PickRequestWithWeight, 0, len(req.FC2.Reqs))
		for _, item := range req.FC2.Reqs {
			converted = append(converted, sc.FC2PickRequestWithWeight{Req: item.FC2Req, Weight: item.Weight})
		}
		result, err := h.scSvc.SmartPickFC2WithInfoFromRequests(c.Request.Context(), converted, req.FC2.PickN, req.FC2.Options)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FC2: " + err.Error()})
			return
		}
		fc2Result = result
	}

	javMovies := make([]*types.MovieType, 0)
	javInfo := sc.SmartPickInfo{}
	if javResult != nil {
		javMovies = javResult.Movies
		javInfo = javResult.Info
	}
	fc2Movies := make([]*sc.FC2SmartPickMovie, 0)
	fc2Info := sc.FC2SmartPickInfo{}
	if fc2Result != nil {
		fc2Movies = fc2Result.Movies
		fc2Info = fc2Result.Info
	}
	fc2MovieRows := make([]*moviex.FC2Movie, 0, len(fc2Movies))
	for _, movie := range fc2Movies {
		if movie != nil && movie.FC2Movie != nil {
			fc2MovieRows = append(fc2MovieRows, movie.FC2Movie)
		}
	}
	fc2CardViews := make([]*mediaservice.FC2CardView, 0, len(fc2MovieRows))
	if len(fc2MovieRows) > 0 {
		views, buildErr := h.mediaSvc.BuildFC2CardViews(c.Request.Context(), fc2MovieRows)
		if buildErr != nil {
			jsonServerError(c, "FC2卡片数据加载失败", buildErr)
			return
		}
		fc2CardViews = views
	}

	javSizeBytes := buildPickedTotalSize(javMovies, source)
	fc2SizeBytes := sc.FC2SmartPickTotalSize(fc2Movies)
	javSizeGB := formatSmartPickSizeGB(javSizeBytes)
	fc2SizeGB := formatSmartPickSizeGB(fc2SizeBytes)
	javInfo.TotalSizeGB = javSizeGB
	fc2Info.TotalSizeGB = fc2SizeGB
	response := gin.H{
		"picked":        len(javMovies) + len(fc2Movies),
		"total_size_gb": formatSmartPickSizeGB(javSizeBytes + fc2SizeBytes),
		"jav": gin.H{
			"picked":        len(javMovies),
			"movies":        buildPickCopyMovies(javMovies, source),
			"total_size_gb": javSizeGB,
			"movie_type":    moviedomain.SehuatangThreadMovieTypeJAV,
			"source":        source,
			"pick_info":     javInfo,
		},
		"fc2": gin.H{
			"picked":        len(fc2Movies),
			"movies":        fc2CardViews,
			"total_size_gb": fc2SizeGB,
			"movie_type":    moviedomain.SehuatangThreadMovieTypeFC2PPV,
			"source":        "fc2media",
			"pick_info":     fc2Info,
		},
	}
	if withCopy {
		started, status := h.scSvc.StartMixedCopyAsync(javMovies, source, fc2Movies)
		response["copy_started"] = started
		response["copy_status"] = status
	}
	c.JSON(http.StatusOK, response)
}

func (h *ScTriggerAPI) CopyStatus(c *gin.Context) {
	status := h.scSvc.CopyStatus()
	c.JSON(http.StatusOK, status)
}

func (h *ScTriggerAPI) CopyStop(c *gin.Context) {
	stopped := h.scSvc.StopCopy()
	status := h.scSvc.CopyStatus()
	c.JSON(http.StatusOK, gin.H{
		"stopped": stopped,
		"status":  status,
	})
}
