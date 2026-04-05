package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/media"
	"rudy_gc/internal/svc"
)

type ScMediaTriggerAPI struct {
	mediaSvc *media.Service
}

func NewScMediaTriggerAPI(deps *svc.Deps) *ScMediaTriggerAPI {
	return &ScMediaTriggerAPI{mediaSvc: media.NewService(deps)}
}

type scMediaMoveReq struct {
	ScName string `json:"sc_name" form:"sc_name"`
}

type scMediaMoveRow struct {
	Status     string `json:"status"`
	MovieJavId string `json:"movie_jav_id"`
	MovieName  string `json:"movie_name"`
	RootDir    string `json:"root_dir"`
	SourcePath string `json:"source_path"`
	TargetDir  string `json:"target_dir"`
	TargetPath string `json:"target_path"`
	Error      string `json:"error"`
}

type scMediaMoveResponse struct {
	Phase         string           `json:"phase"`
	Message       string           `json:"message"`
	ScName        string           `json:"sc_name"`
	HasPlan       bool             `json:"has_plan"`
	GeneratedAt   int64            `json:"generated_at"`
	Total         int              `json:"total"`
	Movable       int              `json:"movable"`
	Skipped       int              `json:"skipped"`
	Failed        int              `json:"failed"`
	Success       int              `json:"success"`
	CommitFailed  int              `json:"commit_failed"`
	CanCommit     bool             `json:"can_commit"`
	PrecheckPass  []scMediaMoveRow `json:"precheck_pass"`
	PrecheckSkip  []scMediaMoveRow `json:"precheck_skip"`
	PrecheckFail  []scMediaMoveRow `json:"precheck_fail"`
	CommitSuccess []scMediaMoveRow `json:"commit_success"`
	CommitFail    []scMediaMoveRow `json:"commit_fail"`
}

func (h *ScMediaTriggerAPI) Plan(c *gin.Context) {
	scName := strings.TrimSpace(c.Query("sc_name"))
	snapshot, err := h.mediaSvc.ReadScMediaMovePlanSnapshot(c.Request.Context(), scName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildScMediaMovePlanResponse(snapshot))
}

func (h *ScMediaTriggerAPI) Precheck(c *gin.Context) {
	var req scMediaMoveReq
	_ = c.ShouldBindJSON(&req)
	req.ScName = strings.TrimSpace(req.ScName)
	if req.ScName == "" {
		req.ScName = strings.TrimSpace(c.Query("sc_name"))
	}
	if req.ScName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sc_name 为空"})
		return
	}

	result, err := h.mediaSvc.ScMediaMovePrecheck(c.Request.Context(), req.ScName)
	resp := buildScMediaMovePrecheckResponse("precheck", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ScMediaTriggerAPI) Commit(c *gin.Context) {
	var req scMediaMoveReq
	_ = c.ShouldBindJSON(&req)
	req.ScName = strings.TrimSpace(req.ScName)
	if req.ScName == "" {
		req.ScName = strings.TrimSpace(c.Query("sc_name"))
	}
	if req.ScName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sc_name 为空"})
		return
	}

	result, err := h.mediaSvc.ScMediaMoveCommit(c.Request.Context(), req.ScName)
	resp := buildScMediaMoveCommitResponse("commit", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ScMediaTriggerAPI) Return(c *gin.Context) {
	var req scMediaMoveReq
	_ = c.ShouldBindJSON(&req)
	req.ScName = strings.TrimSpace(req.ScName)
	if req.ScName == "" {
		req.ScName = strings.TrimSpace(c.Query("sc_name"))
	}
	if req.ScName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sc_name 为空"})
		return
	}
	if err := h.mediaSvc.ClearScMediaMovePlanSnapshot(c.Request.Context(), req.ScName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scMediaMoveResponse{
		Phase:         "return",
		Message:       "已返回，当前预处理计划已清空",
		ScName:        req.ScName,
		HasPlan:       false,
		CanCommit:     false,
		PrecheckPass:  []scMediaMoveRow{},
		PrecheckSkip:  []scMediaMoveRow{},
		PrecheckFail:  []scMediaMoveRow{},
		CommitSuccess: []scMediaMoveRow{},
		CommitFail:    []scMediaMoveRow{},
	})
}

func buildScMediaMovePlanResponse(snapshot *media.ScMediaMovePlanSnapshot) scMediaMoveResponse {
	resp := scMediaMoveResponse{
		Phase:         "plan",
		ScName:        "",
		HasPlan:       false,
		CanCommit:     false,
		PrecheckPass:  []scMediaMoveRow{},
		PrecheckSkip:  []scMediaMoveRow{},
		PrecheckFail:  []scMediaMoveRow{},
		CommitSuccess: []scMediaMoveRow{},
		CommitFail:    []scMediaMoveRow{},
	}
	if snapshot == nil {
		resp.Message = "暂无预处理结果"
		return resp
	}
	resp.ScName = strings.TrimSpace(snapshot.ScName)
	resp.HasPlan = snapshot.HasPlan
	resp.GeneratedAt = snapshot.GeneratedAt
	resp.Total = snapshot.Total
	resp.Movable = snapshot.Movable
	resp.Skipped = snapshot.Skipped
	resp.Failed = snapshot.Failed
	resp.CanCommit = snapshot.HasPlan && snapshot.Movable > 0

	for _, item := range snapshot.Items {
		row := buildScMediaMoveRow(item)
		switch strings.TrimSpace(row.Status) {
		case "pass":
			resp.PrecheckPass = append(resp.PrecheckPass, row)
		case "skip":
			resp.PrecheckSkip = append(resp.PrecheckSkip, row)
		default:
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}

	switch {
	case resp.ScName == "":
		resp.Message = "请先输入 SC 名称"
	case !snapshot.HasPlan:
		resp.Message = "暂无预处理结果，请先执行第一段预处理"
	case snapshot.Total == 0:
		resp.Message = "预处理已完成：当前没有可移动 media"
	case snapshot.Failed == 0 && snapshot.Skipped == 0:
		resp.Message = "预处理已全部通过，可执行第二段移动"
	default:
		resp.Message = "预处理已完成，可执行第二段移动通过项"
	}
	return resp
}

func buildScMediaMovePrecheckResponse(phase string, result *media.ScMediaMoveResult) scMediaMoveResponse {
	resp := scMediaMoveResponse{
		Phase:         phase,
		PrecheckPass:  []scMediaMoveRow{},
		PrecheckSkip:  []scMediaMoveRow{},
		PrecheckFail:  []scMediaMoveRow{},
		CommitSuccess: []scMediaMoveRow{},
		CommitFail:    []scMediaMoveRow{},
	}
	if result == nil {
		resp.Message = "预处理未返回数据"
		return resp
	}

	resp.ScName = strings.TrimSpace(result.ScName)
	resp.HasPlan = true
	resp.GeneratedAt = result.GeneratedAt
	resp.Total = result.Total
	resp.Movable = result.Movable
	resp.Skipped = result.Skipped
	resp.Failed = result.Failed
	resp.CanCommit = result.Movable > 0

	switch {
	case result.Total == 0:
		resp.Message = "预处理完成：当前没有可移动 media"
	case result.Failed == 0 && result.Skipped == 0:
		resp.Message = "预处理完成：全部通过，可执行第二段移动"
	default:
		resp.Message = "预处理完成：存在跳过或失败，第二段仅移动通过项"
	}

	for _, item := range result.Items {
		row := buildScMediaMoveRow(item)
		switch strings.TrimSpace(row.Status) {
		case "pass":
			resp.PrecheckPass = append(resp.PrecheckPass, row)
		case "skip":
			resp.PrecheckSkip = append(resp.PrecheckSkip, row)
		default:
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}
	return resp
}

func buildScMediaMoveCommitResponse(phase string, result *media.ScMediaMoveResult) scMediaMoveResponse {
	resp := scMediaMoveResponse{
		Phase:         phase,
		PrecheckPass:  []scMediaMoveRow{},
		PrecheckSkip:  []scMediaMoveRow{},
		PrecheckFail:  []scMediaMoveRow{},
		CommitSuccess: []scMediaMoveRow{},
		CommitFail:    []scMediaMoveRow{},
	}
	if result == nil {
		resp.Message = "第二段执行未返回数据"
		return resp
	}

	resp.ScName = strings.TrimSpace(result.ScName)
	resp.GeneratedAt = result.GeneratedAt
	resp.Total = result.Total
	resp.Movable = result.Movable
	resp.Skipped = result.Skipped
	resp.Failed = result.Failed
	resp.Success = result.Success
	resp.CommitFailed = result.CommitFailed

	switch {
	case result.Movable == 0:
		resp.Message = "第二段未执行任何移动项"
	case result.CommitFailed == 0:
		resp.Message = "第二段执行完成，全部移动成功"
	default:
		resp.Message = "第二段执行完成，部分移动失败"
	}

	for _, item := range result.Items {
		row := buildScMediaMoveRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.CommitSuccess = append(resp.CommitSuccess, row)
			continue
		}
		resp.CommitFail = append(resp.CommitFail, row)
	}
	return resp
}

func buildScMediaMoveRow(item *media.ScMediaMoveItem) scMediaMoveRow {
	if item == nil {
		return scMediaMoveRow{}
	}
	return scMediaMoveRow{
		Status:     strings.TrimSpace(item.Status),
		MovieJavId: strings.TrimSpace(item.MovieJavId),
		MovieName:  strings.TrimSpace(item.MovieName),
		RootDir:    strings.TrimSpace(item.RootDir),
		SourcePath: strings.TrimSpace(item.SourcePath),
		TargetDir:  strings.TrimSpace(item.TargetDir),
		TargetPath: strings.TrimSpace(item.TargetPath),
		Error:      strings.TrimSpace(item.Error),
	}
}
