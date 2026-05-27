package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/media"
	"rudy_gc/internal/svc"
)

type MediaTriggerAPI struct {
	mediaSvc *media.Service
}

func NewMediaTriggerAPI(deps *svc.Deps) *MediaTriggerAPI {
	return &MediaTriggerAPI{
		mediaSvc: media.NewService(deps),
	}
}

type mediaIngestItemRow struct {
	Status            string `json:"status"`
	MovieName         string `json:"movie_name"`
	SourcePath        string `json:"source_path"`
	TargetFileName    string `json:"target_file_name"`
	TargetDir         string `json:"target_dir"`
	Alias             string `json:"alias"`
	SourceTorrentHash string `json:"source_torrent_hash"`
	Size              int64  `json:"size"`
	BirthTime         int64  `json:"birth_time"`
	TargetPath        string `json:"target_path"`
	FailedPath        string `json:"failed_path"`
	Error             string `json:"error"`
}

type mediaIngestResponse struct {
	Phase           string                   `json:"phase"`
	Message         string                   `json:"message"`
	HasPlan         bool                     `json:"has_plan"`
	GeneratedAt     int64                    `json:"generated_at"`
	Total           int                      `json:"total"`
	Success         int                      `json:"success"`
	Failed          int                      `json:"failed"`
	Precheck        media.IngestPrecheckStat `json:"precheck"`
	CanCommit       bool                     `json:"can_commit"`
	PartialFailed   bool                     `json:"partial_failed"`
	PrecheckPass    []mediaIngestItemRow     `json:"precheck_pass"`
	PrecheckFail    []mediaIngestItemRow     `json:"precheck_fail"`
	PrecheckPreview []mediaIngestItemRow     `json:"precheck_preview"`
	CommitSuccess   []mediaIngestItemRow     `json:"commit_success"`
	CommitFail      []mediaIngestItemRow     `json:"commit_fail"`
	CommitPreview   []mediaIngestItemRow     `json:"commit_preview"`
}

type mediaRollbackResponse struct {
	Phase           string               `json:"phase"`
	Message         string               `json:"message"`
	Total           int                  `json:"total"`
	Success         int                  `json:"success"`
	Failed          int                  `json:"failed"`
	RollbackSuccess []mediaIngestItemRow `json:"rollback_success"`
	RollbackFail    []mediaIngestItemRow `json:"rollback_fail"`
	RollbackPreview []mediaIngestItemRow `json:"rollback_preview"`
}

type mediaRescanRequest struct {
	Selections []media.LibraryRescanSelection `json:"selections"`
}

func (h *MediaTriggerAPI) Precheck(c *gin.Context) {
	result, err := h.mediaSvc.IngestPrecheck(c.Request.Context())
	resp := buildMediaPrecheckResponse("precheck", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MediaTriggerAPI) Commit(c *gin.Context) {
	result, err := h.mediaSvc.IngestCommit(c.Request.Context())
	resp := buildMediaCommitResponse("commit", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MediaTriggerAPI) Plan(c *gin.Context) {
	snapshot, err := h.mediaSvc.ReadIngestPlanSnapshot(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := mediaIngestResponse{
		Phase:       "plan",
		HasPlan:     snapshot.HasPlan,
		GeneratedAt: snapshot.GeneratedAt,
		Total:       snapshot.Total,
		Success:     snapshot.Passed,
		Failed:      snapshot.Failed,
		Precheck: media.IngestPrecheckStat{
			Total:  snapshot.Total,
			Passed: snapshot.Passed,
			Failed: snapshot.Failed,
		},
		CanCommit:       snapshot.HasPlan && snapshot.Passed > 0,
		PartialFailed:   snapshot.HasPlan && snapshot.Failed > 0,
		PrecheckPass:    make([]mediaIngestItemRow, 0, snapshot.Passed),
		PrecheckFail:    make([]mediaIngestItemRow, 0, snapshot.Failed),
		PrecheckPreview: make([]mediaIngestItemRow, 0, snapshot.Passed),
		CommitSuccess:   []mediaIngestItemRow{},
		CommitFail:      []mediaIngestItemRow{},
		CommitPreview:   []mediaIngestItemRow{},
	}

	for _, item := range snapshot.Items {
		row := buildMediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.PrecheckPass = append(resp.PrecheckPass, row)
			resp.PrecheckPreview = append(resp.PrecheckPreview, row)
		} else {
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}

	switch {
	case !snapshot.HasPlan:
		resp.Message = "暂无预处理结果，请先执行第一段预处理"
	case snapshot.Total == 0:
		resp.Message = "预处理已完成：当前没有可处理文件"
	case snapshot.Failed == 0:
		resp.Message = "预处理已全部通过，可执行第二段插入"
	default:
		resp.Message = "预处理存在失败，可选择仅插入通过项或返回"
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MediaTriggerAPI) Return(c *gin.Context) {
	if err := h.mediaSvc.ClearIngestPlanSnapshot(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mediaIngestResponse{
		Phase:           "return",
		Message:         "已返回，当前预处理计划已清空",
		HasPlan:         false,
		Total:           0,
		Success:         0,
		Failed:          0,
		CanCommit:       false,
		PartialFailed:   false,
		PrecheckPass:    []mediaIngestItemRow{},
		PrecheckFail:    []mediaIngestItemRow{},
		PrecheckPreview: []mediaIngestItemRow{},
		CommitSuccess:   []mediaIngestItemRow{},
		CommitFail:      []mediaIngestItemRow{},
		CommitPreview:   []mediaIngestItemRow{},
	})
}

func (h *MediaTriggerAPI) Rollback(c *gin.Context) {
	result, err := h.mediaSvc.RollbackName(c.Request.Context())
	resp := buildMediaRollbackResponse("rollback", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MediaTriggerAPI) Rescan(c *gin.Context) {
	var req mediaRescanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.mediaSvc.RescanLibrary(c.Request.Context(), req.Selections)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"result": result,
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func buildMediaPrecheckResponse(phase string, result *media.IngestNewResult) mediaIngestResponse {
	resp := mediaIngestResponse{
		Phase:           phase,
		HasPlan:         true,
		PrecheckPass:    []mediaIngestItemRow{},
		PrecheckFail:    []mediaIngestItemRow{},
		PrecheckPreview: []mediaIngestItemRow{},
		CommitSuccess:   []mediaIngestItemRow{},
		CommitFail:      []mediaIngestItemRow{},
		CommitPreview:   []mediaIngestItemRow{},
	}
	if result == nil {
		resp.HasPlan = false
		resp.Message = "预处理未返回数据"
		return resp
	}

	resp.Precheck = result.Precheck
	resp.Total = result.Precheck.Total
	resp.Success = result.Precheck.Passed
	resp.Failed = result.Precheck.Failed
	resp.CanCommit = result.Precheck.Passed > 0
	resp.PartialFailed = result.Precheck.Failed > 0
	if result.Precheck.Total == 0 {
		resp.Message = "预处理完成：没有可处理文件"
	} else if result.Precheck.Failed == 0 {
		resp.Message = "预处理完成：全部通过，可执行第二段插入"
	} else {
		resp.Message = "预处理完成：部分失败，可仅插入通过项或返回"
	}

	for _, item := range result.Items {
		row := buildMediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.PrecheckPass = append(resp.PrecheckPass, row)
			resp.PrecheckPreview = append(resp.PrecheckPreview, row)
		} else {
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}
	return resp
}

func buildMediaCommitResponse(phase string, result *media.IngestNewResult) mediaIngestResponse {
	resp := mediaIngestResponse{
		Phase:           phase,
		HasPlan:         false,
		CanCommit:       false,
		PartialFailed:   false,
		PrecheckPass:    []mediaIngestItemRow{},
		PrecheckFail:    []mediaIngestItemRow{},
		PrecheckPreview: []mediaIngestItemRow{},
		CommitSuccess:   []mediaIngestItemRow{},
		CommitFail:      []mediaIngestItemRow{},
		CommitPreview:   []mediaIngestItemRow{},
	}
	if result == nil {
		resp.Message = "第二段执行未返回数据"
		return resp
	}

	resp.Precheck = result.Precheck
	resp.Total = result.Total
	resp.Success = result.Success
	resp.Failed = result.Failed
	if result.Total == 0 {
		resp.Message = "第二段未执行任何插入项"
	} else if result.Failed == 0 {
		resp.Message = "第二段执行完成，全部插入成功"
	} else {
		resp.Message = "第二段执行完成，部分插入失败"
	}

	for _, item := range result.Items {
		row := buildMediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.CommitSuccess = append(resp.CommitSuccess, row)
			resp.CommitPreview = append(resp.CommitPreview, row)
		} else {
			resp.CommitFail = append(resp.CommitFail, row)
		}
	}
	return resp
}

func buildMediaRollbackResponse(phase string, result *media.IngestNewResult) mediaRollbackResponse {
	resp := mediaRollbackResponse{
		Phase:           phase,
		RollbackSuccess: []mediaIngestItemRow{},
		RollbackFail:    []mediaIngestItemRow{},
		RollbackPreview: []mediaIngestItemRow{},
	}
	if result == nil {
		resp.Message = "回滚未返回数据"
		return resp
	}

	resp.Total = result.Total
	resp.Success = result.Success
	resp.Failed = result.Failed
	switch {
	case result.Total == 0:
		resp.Message = "回滚完成：没有可处理文件"
	case result.Failed == 0:
		resp.Message = "回滚完成：全部成功"
	default:
		resp.Message = "回滚完成：部分失败"
	}

	for _, item := range result.Items {
		row := buildMediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.RollbackSuccess = append(resp.RollbackSuccess, row)
			resp.RollbackPreview = append(resp.RollbackPreview, row)
			continue
		}
		resp.RollbackFail = append(resp.RollbackFail, row)
	}
	return resp
}

func buildMediaIngestRow(item *media.IngestFileItem) mediaIngestItemRow {
	if item == nil {
		return mediaIngestItemRow{}
	}
	return mediaIngestItemRow{
		Status:            strings.TrimSpace(item.Status),
		MovieName:         strings.TrimSpace(item.MovieName),
		SourcePath:        strings.TrimSpace(item.SourcePath),
		TargetFileName:    strings.TrimSpace(item.TargetFileName),
		TargetDir:         strings.TrimSpace(item.TargetDir),
		Alias:             strings.TrimSpace(item.Alias),
		SourceTorrentHash: strings.TrimSpace(item.SourceTorrentHash),
		Size:              item.Size,
		BirthTime:         item.BirthTime,
		TargetPath:        strings.TrimSpace(item.TargetPath),
		FailedPath:        strings.TrimSpace(item.FailedPath),
		Error:             strings.TrimSpace(item.Error),
	}
}
