package movie

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	personservice "rudy_gc/local/api/internal/service/person"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
)

type PersonAPI struct {
	scSvc     *sc.Service
	personSvc *personservice.Service
}

func NewPersonAPI(deps *dep.Dep) *PersonAPI {
	return &PersonAPI{
		scSvc:     sc.NewService(deps),
		personSvc: personservice.NewService(deps),
	}
}

func (h *PersonAPI) ListCafoProfiles(c *gin.Context) {
	personID, ok := parsePersonID(c)
	if !ok {
		return
	}
	rows, err := h.personSvc.ListCafoProfiles(c.Request.Context(), personID)
	if err != nil {
		writeCafoProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "cafos": rows})
}

func (h *PersonAPI) UpdateCafoProfile(c *gin.Context) {
	personID, ok := parsePersonID(c)
	if !ok {
		return
	}
	cafoID, err := strconv.ParseInt(strings.TrimSpace(c.Param("cafo_id")), 10, 64)
	if err != nil || cafoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "cafo id 无效"})
		return
	}
	var req types.CafoProfileWriteForm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}
	cafo, person, err := h.personSvc.UpdateCafoProfile(c.Request.Context(), personID, cafoID, &req)
	if err != nil {
		writeCafoProfileError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "演员资料更新成功",
		"cafo":    cafo,
		"person":  person,
	})
}

func (h *PersonAPI) MergeCandidates(c *gin.Context) {
	keepPersonID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || keepPersonID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "person id 无效"})
		return
	}

	keyword := strings.TrimSpace(c.Query("q"))
	limit := parsePositiveInt64OrDefault(c.Query("limit"), 12)

	rows, err := h.scSvc.SearchPersonMergeCandidates(c.Request.Context(), keepPersonID, keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"candidates": rows,
	})
}

func (h *PersonAPI) MergePreview(c *gin.Context) {
	var req personMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	preview, err := h.scSvc.PreviewPersonMerge(c.Request.Context(), req.KeepPersonId, req.SourcePersonIds)
	if err != nil {
		writePersonMergeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"preview": preview,
	})
}

func (h *PersonAPI) Merge(c *gin.Context) {
	var req personMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	result, err := h.scSvc.MergePerson(c.Request.Context(), req.KeepPersonId, req.SourcePersonIds)
	if err != nil {
		writePersonMergeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":     true,
		"result": result,
	})
}

type personMergeRequest struct {
	KeepPersonId    int64   `json:"keepPersonId"`
	SourcePersonIds []int64 `json:"sourcePersonIds"`
}

func writePersonMergeError(c *gin.Context, err error) {
	if errors.Is(err, types.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "person 不存在"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
}

func parsePersonID(c *gin.Context) (int64, bool) {
	personID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || personID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "person id 无效"})
		return 0, false
	}
	return personID, true
}

func writeCafoProfileError(c *gin.Context, err error) {
	if errors.Is(err, types.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "人物或关联的 c_cafo 不存在"})
		return
	}
	var badRequest *types.BadRequestError
	if errors.As(err, &badRequest) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": badRequest.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
}

func parsePositiveInt64OrDefault(raw string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}
