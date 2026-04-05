package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/filmmove"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type FilmMoveAPI struct {
	svc *filmmove.Service
}

func NewFilmMoveAPI(deps *svc.Deps) *FilmMoveAPI {
	return &FilmMoveAPI{svc: filmmove.NewService(deps)}
}

type filmMoveItemRow struct {
	MovieName  string `json:"movie_name"`
	MovieJavID string `json:"movie_jav_id"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	CanMove    bool   `json:"can_move"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

type filmMovePreviewResp struct {
	Phase   string            `json:"phase"`
	Message string            `json:"message"`
	PlanID  string            `json:"plan_id"`
	Total   int               `json:"total"`
	Movable int               `json:"movable"`
	Failed  int               `json:"failed"`
	Items   []filmMoveItemRow `json:"items"`
}

type filmMoveCommitReq struct {
	PlanID string `json:"plan_id"`
}

type filmMoveCommitResp struct {
	Phase        string            `json:"phase"`
	Message      string            `json:"message"`
	PlanID       string            `json:"plan_id"`
	Total        int               `json:"total"`
	Success      int               `json:"success"`
	Failed       int               `json:"failed"`
	Items        []filmMoveItemRow `json:"items"`
	SuccessItems []filmMoveItemRow `json:"success_items"`
	FailedItems  []filmMoveItemRow `json:"failed_items"`
}

func (h *FilmMoveAPI) Preview(c *gin.Context) {
	req, err := parseMovieCardRequest(c, types.ListMovieFullRequest{
		OrderBy:  consts.OrderByReleasingDate,
		Page:     1,
		PageSize: maxPageSize,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	result, err := h.svc.Preview(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := filmMovePreviewResp{
		Phase:   "preview",
		PlanID:  result.PlanID,
		Total:   result.Total,
		Movable: result.Movable,
		Failed:  result.Failed,
		Items:   make([]filmMoveItemRow, 0, len(result.Items)),
	}
	switch {
	case result.Total == 0:
		resp.Message = "第一步完成：未查到影片"
	case result.Movable == 0:
		resp.Message = "第一步完成：有结果但均不可移动"
	case result.Failed == 0:
		resp.Message = "第一步完成：全部可移动，可执行第二步"
	default:
		resp.Message = "第一步完成：部分可移动，可执行第二步"
	}

	for _, item := range result.Items {
		if item == nil {
			continue
		}
		resp.Items = append(resp.Items, filmMoveItemRow{
			MovieName:  strings.TrimSpace(item.MovieName),
			MovieJavID: strings.TrimSpace(item.MovieJavID),
			SourcePath: strings.TrimSpace(item.SourcePath),
			TargetPath: strings.TrimSpace(item.TargetPath),
			CanMove:    item.CanMove,
			Status:     boolToMoveStatus(item.CanMove),
			Error:      strings.TrimSpace(item.Error),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FilmMoveAPI) Commit(c *gin.Context) {
	var req filmMoveCommitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误"})
		return
	}

	result, err := h.svc.Commit(c.Request.Context(), req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := filmMoveCommitResp{
		Phase:        "commit",
		PlanID:       strings.TrimSpace(result.PlanID),
		Total:        result.Total,
		Success:      result.Success,
		Failed:       result.Failed,
		Items:        make([]filmMoveItemRow, 0, len(result.Items)),
		SuccessItems: make([]filmMoveItemRow, 0, len(result.SuccessItems)),
		FailedItems:  make([]filmMoveItemRow, 0, len(result.FailedItems)),
	}
	switch {
	case result.Total == 0:
		resp.Message = "第二步完成：没有待移动影片"
	case result.Failed == 0:
		resp.Message = "第二步完成：全部移动成功"
	case result.Success == 0:
		resp.Message = "第二步完成：全部移动失败"
	default:
		resp.Message = "第二步完成：部分移动成功"
	}

	for _, item := range result.Items {
		if item == nil {
			continue
		}
		resp.Items = append(resp.Items, toFilmMoveRow(item))
	}
	for _, item := range result.SuccessItems {
		if item == nil {
			continue
		}
		resp.SuccessItems = append(resp.SuccessItems, toFilmMoveRow(item))
	}
	for _, item := range result.FailedItems {
		if item == nil {
			continue
		}
		resp.FailedItems = append(resp.FailedItems, toFilmMoveRow(item))
	}

	c.JSON(http.StatusOK, resp)
}

func toFilmMoveRow(item *filmmove.CommitItem) filmMoveItemRow {
	return filmMoveItemRow{
		MovieName:  strings.TrimSpace(item.MovieName),
		MovieJavID: strings.TrimSpace(item.MovieJavID),
		SourcePath: strings.TrimSpace(item.SourcePath),
		TargetPath: strings.TrimSpace(item.TargetPath),
		CanMove:    strings.TrimSpace(item.Status) == "success",
		Status:     strings.TrimSpace(item.Status),
		Error:      strings.TrimSpace(item.Error),
	}
}

func boolToMoveStatus(canMove bool) string {
	if canMove {
		return "ready"
	}
	return "blocked"
}
