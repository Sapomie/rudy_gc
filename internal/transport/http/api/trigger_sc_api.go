// internal/transport/http/api/trigger_api.go（或新文件 sc_trigger_api.go）
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/contracts"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type ScTriggerAPI struct {
	ch    chan contracts.ScTriggerMsg
	scSvc *sc.ScService
}

func NewScTriggerAPI(deps *svc.Deps) *ScTriggerAPI {
	return &ScTriggerAPI{
		ch:    deps.ScTrigger,
		scSvc: sc.NewScService(deps),
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

// POST /api/triggers/sc/add { "dir": "/path/to/scan" }
type scAddReq struct {
	Dir string `json:"dir" form:"dir"`
}

func (h *ScTriggerAPI) Add(c *gin.Context) {
	var req scAddReq
	_ = c.ShouldBind(&req)
	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dir required"})
		return
	}
	msg := contracts.ScTriggerMsg{Kind: contracts.ScAdd, Dir: dir}
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
	c.JSON(http.StatusOK, gin.H{
		"picked": len(movies),
		"movies": buildPickCopyMovies(movies),
	})
}
