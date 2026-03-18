package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/svc"
)

type FilmTriggerAPI struct {
	runtime *loop.FetchLoopService
}

func NewFilmTriggerAPI(deps *svc.Deps) *FilmTriggerAPI {
	return &FilmTriggerAPI{runtime: newCrawlerRuntime(deps)}
}

// POST /api/triggers/film/rename
func (h *FilmTriggerAPI) Rename(c *gin.Context) {
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskFilmRename})
}

// POST /api/triggers/film/process
func (h *FilmTriggerAPI) Process(c *gin.Context) {
	h.startTask(c, loop.StartTaskRequest{TaskType: loop.TaskFilmProcess})
}

func (h *FilmTriggerAPI) startTask(c *gin.Context, req loop.StartTaskRequest) {
	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobStarted(c, jobID, req.TaskType)
}
