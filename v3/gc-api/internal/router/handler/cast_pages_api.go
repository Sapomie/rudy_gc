package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/types"
)

type CastPagesAPI struct {
	scSvc *sc.ScService
	cards *MovieCardAPI
}

func NewCastPagesAPI(deps *dep.Dep) *CastPagesAPI {
	return &CastPagesAPI{
		scSvc: sc.NewService(deps),
		cards: NewMovieCardAPI(deps),
	}
}

func (h *CastPagesAPI) Detail(c *gin.Context) {
	personID, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	name := strings.TrimSpace(c.Query("name"))

	var (
		vm  *types.ActorScPage
		err error
	)
	if personID > 0 {
		vm, err = h.scSvc.BuildActorScPageByPersonID(c.Request.Context(), personID, 0, 24)
	} else {
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: id 或 name"})
			return
		}
		vm, err = h.scSvc.BuildActorScPage(c.Request.Context(), name, 0, 24)
	}
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			if personID > 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "未找到演员: " + strconv.FormatInt(personID, 10)})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到演员: " + name})
			return
		}
		jsonServerError(c, "演员页加载失败", err)
		return
	}
	if vm == nil || vm.Actor == nil {
		if personID > 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到演员: " + strconv.FormatInt(personID, 10)})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到演员: " + name})
		return
	}

	cardReq, err := parseMovieCardRequest(c, types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate})
	if err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}
	if vm.Actor.Id > 0 {
		cardReq.PersonIds = strconv.FormatInt(vm.Actor.Id, 10)
		cardReq.CastNames = ""
	} else {
		cardReq.PersonIds = ""
		cardReq.CastNames = strings.TrimSpace(vm.Actor.Name)
	}

	movieCardAction, movieCardClearHref := buildCastDetailMovieCardTargets(vm.Actor, name)
	frontendPath := buildCastDetailFrontendPath(vm.Actor, name)
	restorePath := withFrontendPath(c, frontendPath)
	defer restorePath()

	movieCardData, err := h.cards.loadMovieCardPageData(c, cardReq, movieCardAction, movieCardClearHref)
	if err != nil {
		jsonServerError(c, "演员影片列表加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":           personDisplayName(vm.Actor),
		"ViewType":        "cast-detail",
		"Page":            vm,
		"movies":          movieCardData.Movies,
		"total":           movieCardData.Total,
		"PageInfo":        movieCardData.PageInfo,
		"pageInfo":        movieCardData.PageInfo,
		"ownedQuery":      movieCardData.OwnedQuery,
		"sortQuery":       movieCardData.SortQuery,
		"CurrentSort":     movieCardData.CurrentSort,
		"MovieCardFilter": movieCardData.MovieCardFilter,
	})
}

func personDisplayName(person *types.Person) string {
	if person == nil {
		return ""
	}
	if name := strings.TrimSpace(person.Chinese); name != "" {
		return name
	}
	return strings.TrimSpace(person.Name)
}

func buildCastDetailFrontendPath(person *types.Person, rawName string) string {
	if person != nil && person.Id > 0 {
		return "/castdetail/" + strconv.FormatInt(person.Id, 10)
	}
	return "/cast"
}

func buildCastDetailMovieCardTargets(person *types.Person, rawName string) (string, string) {
	if person != nil && person.Id > 0 {
		target := "/castdetail/" + strconv.FormatInt(person.Id, 10)
		return target, target
	}

	values := url.Values{}
	name := strings.TrimSpace(rawName)
	if name == "" && person != nil {
		name = strings.TrimSpace(person.Name)
	}
	if name != "" {
		values.Set("name", name)
	}
	path := "/cast"
	if encoded := values.Encode(); encoded != "" {
		target := path + "?" + encoded
		return target, target
	}
	return path, path
}
