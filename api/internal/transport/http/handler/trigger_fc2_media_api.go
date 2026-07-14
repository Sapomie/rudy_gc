package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/media"
)

type FC2MediaTriggerAPI struct {
	mediaSvc *media.Service
}

func NewFC2MediaTriggerAPI(deps *dep.Dep) *FC2MediaTriggerAPI {
	return &FC2MediaTriggerAPI{
		mediaSvc: media.NewService(deps),
	}
}

type fc2MediaIngestItemRow struct {
	Status            string `json:"status"`
	FC2Code           string `json:"fc2_code"`
	FC2Key            string `json:"fc2_key"`
	FC2MovieID        int64  `json:"fc2_movie_id"`
	FC2MediaID        int64  `json:"fc2_media_id"`
	SourceThreadID    int64  `json:"source_thread_id"`
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

type fc2MediaIngestResponse struct {
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
	PrecheckPass    []fc2MediaIngestItemRow  `json:"precheck_pass"`
	PrecheckFail    []fc2MediaIngestItemRow  `json:"precheck_fail"`
	PrecheckPreview []fc2MediaIngestItemRow  `json:"precheck_preview"`
	CommitSuccess   []fc2MediaIngestItemRow  `json:"commit_success"`
	CommitFail      []fc2MediaIngestItemRow  `json:"commit_fail"`
	CommitPreview   []fc2MediaIngestItemRow  `json:"commit_preview"`
}

func (h *FC2MediaTriggerAPI) Precheck(c *gin.Context) {
	result, err := h.mediaSvc.FC2IngestPrecheck(c.Request.Context())
	resp := buildFC2MediaPrecheckResponse("precheck", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "result": resp})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FC2MediaTriggerAPI) Commit(c *gin.Context) {
	result, err := h.mediaSvc.FC2IngestCommit(c.Request.Context())
	resp := buildFC2MediaCommitResponse("commit", result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "result": resp})
		return
	}
	c.JSON(http.StatusOK, resp)
	h.mediaSvc.EnqueueFC2DefaultCovers(successfulFC2MediaIDs(result))
}

func (h *FC2MediaTriggerAPI) BackfillDefaultCovers(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"ok": true, "message": "FC2默认联系表封面补齐任务已启动"})
	h.mediaSvc.EnqueueFC2DefaultCoverBackfill()
}

func successfulFC2MediaIDs(result *media.FC2IngestResult) []int64 {
	if result == nil {
		return nil
	}
	mediaIDs := make([]int64, 0, result.Success)
	for _, item := range result.Items {
		if item != nil && strings.TrimSpace(item.Error) == "" && item.FC2MediaID > 0 {
			mediaIDs = append(mediaIDs, item.FC2MediaID)
		}
	}
	return mediaIDs
}

func (h *FC2MediaTriggerAPI) Plan(c *gin.Context) {
	snapshot, err := h.mediaSvc.ReadFC2IngestPlanSnapshot(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := fc2MediaIngestResponse{
		Phase:           "plan",
		HasPlan:         snapshot.HasPlan,
		GeneratedAt:     snapshot.GeneratedAt,
		Total:           snapshot.Total,
		Success:         snapshot.Passed,
		Failed:          snapshot.Failed,
		Precheck:        media.IngestPrecheckStat{Total: snapshot.Total, Passed: snapshot.Passed, Failed: snapshot.Failed},
		CanCommit:       snapshot.HasPlan && snapshot.Passed > 0,
		PartialFailed:   snapshot.HasPlan && snapshot.Failed > 0,
		PrecheckPass:    make([]fc2MediaIngestItemRow, 0, snapshot.Passed),
		PrecheckFail:    make([]fc2MediaIngestItemRow, 0, snapshot.Failed),
		PrecheckPreview: make([]fc2MediaIngestItemRow, 0, snapshot.Passed),
		CommitSuccess:   []fc2MediaIngestItemRow{},
		CommitFail:      []fc2MediaIngestItemRow{},
		CommitPreview:   []fc2MediaIngestItemRow{},
	}
	for _, item := range snapshot.Items {
		row := buildFC2MediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.PrecheckPass = append(resp.PrecheckPass, row)
			resp.PrecheckPreview = append(resp.PrecheckPreview, row)
		} else {
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}
	resp.Message = buildFC2PlanMessage(snapshot.HasPlan, snapshot.Total, snapshot.Failed)
	c.JSON(http.StatusOK, resp)
}

func (h *FC2MediaTriggerAPI) Return(c *gin.Context) {
	if err := h.mediaSvc.ClearFC2IngestPlan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fc2MediaIngestResponse{
		Phase:           "return",
		Message:         "已返回，当前预处理计划已清空",
		HasPlan:         false,
		Total:           0,
		Success:         0,
		Failed:          0,
		CanCommit:       false,
		PartialFailed:   false,
		PrecheckPass:    []fc2MediaIngestItemRow{},
		PrecheckFail:    []fc2MediaIngestItemRow{},
		PrecheckPreview: []fc2MediaIngestItemRow{},
		CommitSuccess:   []fc2MediaIngestItemRow{},
		CommitFail:      []fc2MediaIngestItemRow{},
		CommitPreview:   []fc2MediaIngestItemRow{},
	})
}

func buildFC2MediaPrecheckResponse(phase string, result *media.FC2IngestResult) fc2MediaIngestResponse {
	resp := fc2MediaIngestResponse{
		Phase:           phase,
		HasPlan:         true,
		PrecheckPass:    []fc2MediaIngestItemRow{},
		PrecheckFail:    []fc2MediaIngestItemRow{},
		PrecheckPreview: []fc2MediaIngestItemRow{},
		CommitSuccess:   []fc2MediaIngestItemRow{},
		CommitFail:      []fc2MediaIngestItemRow{},
		CommitPreview:   []fc2MediaIngestItemRow{},
	}
	if result == nil {
		resp.HasPlan = false
		resp.Message = "预处理未返回数据"
		return resp
	}
	resp.GeneratedAt = result.GeneratedAt
	resp.Precheck = result.Precheck
	resp.Total = result.Precheck.Total
	resp.Success = result.Precheck.Passed
	resp.Failed = result.Precheck.Failed
	resp.CanCommit = result.Precheck.Passed > 0
	resp.PartialFailed = result.Precheck.Failed > 0
	switch {
	case result.Precheck.Total == 0:
		resp.Message = "预处理完成：没有可处理文件"
	case result.Precheck.Failed == 0:
		resp.Message = "预处理完成：全部通过，可执行第二段插入"
	default:
		resp.Message = "预处理完成：部分失败，可仅插入通过项或返回"
	}
	for _, item := range result.Items {
		row := buildFC2MediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.PrecheckPass = append(resp.PrecheckPass, row)
			resp.PrecheckPreview = append(resp.PrecheckPreview, row)
		} else {
			resp.PrecheckFail = append(resp.PrecheckFail, row)
		}
	}
	return resp
}

func buildFC2MediaCommitResponse(phase string, result *media.FC2IngestResult) fc2MediaIngestResponse {
	resp := fc2MediaIngestResponse{
		Phase:           phase,
		HasPlan:         false,
		CanCommit:       false,
		PartialFailed:   false,
		PrecheckPass:    []fc2MediaIngestItemRow{},
		PrecheckFail:    []fc2MediaIngestItemRow{},
		PrecheckPreview: []fc2MediaIngestItemRow{},
		CommitSuccess:   []fc2MediaIngestItemRow{},
		CommitFail:      []fc2MediaIngestItemRow{},
		CommitPreview:   []fc2MediaIngestItemRow{},
	}
	if result == nil {
		resp.Message = "第二段执行未返回数据"
		return resp
	}
	resp.Precheck = result.Precheck
	resp.Total = result.Total
	resp.Success = result.Success
	resp.Failed = result.Failed
	switch {
	case result.Total == 0:
		resp.Message = "第二段未执行任何插入项"
	case result.Failed == 0:
		resp.Message = "第二段执行完成，全部插入成功"
	default:
		resp.Message = "第二段执行完成，部分插入失败"
	}
	for _, item := range result.Items {
		row := buildFC2MediaIngestRow(item)
		if strings.TrimSpace(row.Error) == "" {
			resp.CommitSuccess = append(resp.CommitSuccess, row)
			resp.CommitPreview = append(resp.CommitPreview, row)
		} else {
			resp.CommitFail = append(resp.CommitFail, row)
		}
	}
	return resp
}

func buildFC2PlanMessage(hasPlan bool, total, failed int) string {
	switch {
	case !hasPlan:
		return "暂无预处理结果，请先执行第一段预处理"
	case total == 0:
		return "预处理已完成：当前没有可处理文件"
	case failed == 0:
		return "预处理已全部通过，可执行第二段插入"
	default:
		return "预处理存在失败，可选择仅插入通过项或返回"
	}
}

func buildFC2MediaIngestRow(item *media.FC2IngestFileItem) fc2MediaIngestItemRow {
	if item == nil {
		return fc2MediaIngestItemRow{}
	}
	return fc2MediaIngestItemRow{
		Status:            strings.TrimSpace(item.Status),
		FC2Code:           strings.TrimSpace(item.FC2Code),
		FC2Key:            strings.TrimSpace(item.FC2Key),
		FC2MovieID:        item.FC2MovieID,
		FC2MediaID:        item.FC2MediaID,
		SourceThreadID:    item.SourceThreadID,
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
