package sc

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/router/handler/common"
	mediaservice "rudy_gc/local/api/internal/service/media"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type scPickV2Cast struct {
	PersonID int64  `json:"person_id"`
	Name     string `json:"name"`
}

type scPickV2Movie struct {
	Name                 string         `json:"name"`
	Title                string         `json:"title"`
	ReleasingDate        string         `json:"releasing_date"`
	FilmBirthDateWMedia  string         `json:"film_birth_date_w_media"`
	Score                float64        `json:"score"`
	ViewersNumberWatched int64          `json:"viewers_number_watched"`
	Director             string         `json:"director"`
	Genre                []string       `json:"genre"`
	Cast                 []scPickV2Cast `json:"cast"`
	JavURL               string         `json:"jav_url"`
	VideoURLWMedia       string         `json:"video_url_w_media"`
	SearchURL            string         `json:"search_url"`
	BusURL               string         `json:"bus_url"`
	JacketImage          string         `json:"jacket_img"`
	OwnedWMedia          int64          `json:"owned_w_media"`
	Prefix               string         `json:"prefix"`
	ScTimes              int64          `json:"sc_times"`
	ComeTimes            int64          `json:"come_times"`
	HighestRank          int64          `json:"highest_rank"`
	PickedOrder          int            `json:"picked_order"`
}

type scPickV2CopyReq struct {
	PreviewID string `json:"preview_id"`
}

func (h *ScTriggerAPI) PickSmartV2(c *gin.Context) {
	var req scSmartPickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := validateScSmartPickReq(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source := sc.NormalizeSmartPickSource(req.Source)
	javMovies := make([]*types.MovieType, 0)
	javInfo := sc.SmartPickInfo{}
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
		javMovies = result.Movies
		javInfo = result.Info
	}

	fc2Movies := make([]*sc.FC2SmartPickMovie, 0)
	fc2Info := sc.FC2SmartPickInfo{}
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
		fc2Movies = result.Movies
		fc2Info = result.Info
	}

	fc2MovieRows := make([]*moviex.FC2Movie, 0, len(fc2Movies))
	for _, movie := range fc2Movies {
		if movie != nil && movie.FC2Movie != nil {
			fc2MovieRows = append(fc2MovieRows, movie.FC2Movie)
		}
	}
	fc2Cards := make([]*mediaservice.FC2CardView, 0, len(fc2MovieRows))
	if len(fc2MovieRows) > 0 {
		views, err := h.mediaSvc.BuildFC2CardViews(c.Request.Context(), fc2MovieRows)
		if err != nil {
			common.JSONServerError(c, "FC2卡片数据加载失败", err)
			return
		}
		fc2Cards = views
	}

	previewID, err := h.scSvc.CreateMixedCopyPreview(javMovies, source, fc2Movies)
	if err != nil {
		common.JSONServerError(c, "创建复制预览失败", err)
		return
	}
	javBytes := buildPickedTotalSize(javMovies, source)
	fc2Bytes := sc.FC2SmartPickTotalSize(fc2Movies)
	javInfo.TotalSizeGB = formatSmartPickSizeGB(javBytes)
	fc2Info.TotalSizeGB = formatSmartPickSizeGB(fc2Bytes)
	c.JSON(http.StatusOK, gin.H{
		"preview_id":       previewID,
		"picked":           len(javMovies) + len(fc2Movies),
		"total_size_bytes": javBytes + fc2Bytes,
		"total_size_gb":    formatSmartPickSizeGB(javBytes + fc2Bytes),
		"jav": gin.H{
			"picked":           len(javMovies),
			"movies":           buildPickV2Movies(javMovies, source),
			"total_size_bytes": javBytes,
			"total_size_gb":    javInfo.TotalSizeGB,
			"movie_type":       moviedomain.SehuatangThreadMovieTypeJAV,
			"source":           source,
			"pick_info":        javInfo,
		},
		"fc2": gin.H{
			"picked":           len(fc2Movies),
			"movies":           fc2Cards,
			"total_size_bytes": fc2Bytes,
			"total_size_gb":    fc2Info.TotalSizeGB,
			"movie_type":       moviedomain.SehuatangThreadMovieTypeFC2PPV,
			"source":           "fc2media",
			"pick_info":        fc2Info,
		},
	})
}

func (h *ScTriggerAPI) StartCopyV2(c *gin.Context) {
	var req scPickV2CopyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	status, err := h.scSvc.StartCopyV2(req.PreviewID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, status)
}

func (h *ScTriggerAPI) CopyStatusV2(c *gin.Context) {
	status, ok := h.scSvc.CopyV2Status(c.Param("job_id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "copy job not found"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *ScTriggerAPI) StopCopyV2(c *gin.Context) {
	status, ok := h.scSvc.StopCopyV2(c.Param("job_id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "copy job not found"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func buildPickV2Movies(movies []*types.MovieType, source string) []scPickV2Movie {
	if len(movies) == 0 {
		return []scPickV2Movie{}
	}
	result := make([]scPickV2Movie, 0, len(movies))
	for index, movie := range movies {
		if movie == nil {
			continue
		}
		casts := make([]scPickV2Cast, 0, len(movie.Cast))
		for _, cast := range movie.Cast {
			if cast == nil {
				continue
			}
			name := strings.TrimSpace(cast.NameShow)
			if name == "" {
				name = strings.TrimSpace(cast.Name)
			}
			casts = append(casts, scPickV2Cast{PersonID: cast.PersonId, Name: name})
		}
		result = append(result, scPickV2Movie{
			Name:                 movie.Name,
			Title:                movie.Title,
			ReleasingDate:        movie.ReleasingDate,
			FilmBirthDateWMedia:  sc.SmartPickMovieBirthDate(movie, source),
			Score:                movie.Score,
			ViewersNumberWatched: movie.ViewersNumberWatched,
			Director:             movie.Director,
			Genre:                movie.Genre,
			Cast:                 casts,
			JavURL:               movie.JavUrl,
			VideoURLWMedia:       sc.SmartPickMovieVideoURL(movie, source),
			SearchURL:            movie.SearchUrl,
			BusURL:               movie.BusUrl,
			JacketImage:          movie.JacketImg,
			OwnedWMedia:          sc.SmartPickMovieOwned(movie, source),
			Prefix:               movie.Prefix,
			ScTimes:              movie.ScTimes,
			ComeTimes:            movie.ComeTimes,
			HighestRank:          movie.HighestRank,
			PickedOrder:          index,
		})
	}
	return result
}
