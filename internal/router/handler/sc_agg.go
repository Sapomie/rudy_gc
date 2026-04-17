package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (h *MovieHTMLHandler) CastDetailPage(c *gin.Context) {
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
			c.String(http.StatusBadRequest, "缺少参数: id 或 name")
			return
		}
		vm, err = h.scSvc.BuildActorScPage(c.Request.Context(), name, 0, 24)
	}
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			if personID > 0 {
				c.String(http.StatusNotFound, "未找到演员: %d", personID)
				return
			}
			c.String(http.StatusNotFound, "未找到演员: %s", name)
			return
		}
		c.String(http.StatusInternalServerError, "演员页加载失败: %v", err)
		return
	}
	if vm == nil || vm.Actor == nil {
		if personID > 0 {
			c.String(http.StatusNotFound, "未找到演员: %d", personID)
			return
		}
		c.String(http.StatusNotFound, "未找到演员: %s", name)
		return
	}

	cardReq, err := parseMovieCardRequest(c, types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate})
	if err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
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
	movieCardData, err := h.loadMovieCardPageData(c, cardReq, movieCardAction, movieCardClearHref)
	if err != nil {
		c.String(http.StatusInternalServerError, "演员影片列表加载失败: %v", err)
		return
	}

	c.HTML(http.StatusOK, "page.cast_detail", gin.H{
		"Title":           personDisplayName(vm.Actor),
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

func buildCastDetailMovieCardTargets(person *types.Person, rawName string) (string, string) {
	if person != nil && person.Id > 0 {
		target := "/cast/" + strconv.FormatInt(person.Id, 10)
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
