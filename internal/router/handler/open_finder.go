//go:build darwin

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type openReq struct {
	Path string `json:"path"`
}

func (h *MovieAPI) OpenFinderHandler(allowRoots []string) gin.HandlerFunc {
	for i, p := range allowRoots {
		abs, _ := filepath.Abs(p)
		allowRoots[i] = abs
	}
	return func(c *gin.Context) {
		var req openReq
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			c.String(http.StatusBadRequest, "bad json")
			return
		}
		fmt.Println("11122222")

		if req.Path == "" {
			c.String(http.StatusBadRequest, "missing path")
			return
		}
		target, err := secureCleanAbs(req.Path)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid path")
			return
		}
		if !underAnyRoot(target, allowRoots) {
			c.String(http.StatusForbidden, "path not allowed: "+html.EscapeString(target))
			return
		}
		fi, err := os.Stat(target)
		if err != nil {
			c.String(http.StatusNotFound, "not found: "+html.EscapeString(err.Error()))
			return
		}

		// 同步执行并返回错误信息
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var cmd *exec.Cmd
		if fi.IsDir() {
			cmd = exec.CommandContext(ctx, "open", target)
		} else {
			cmd = exec.CommandContext(ctx, "open", "-R", target) // 在 Finder 中定位文件
		}
		out, err := cmd.CombinedOutput() // 等待执行完成
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			c.String(http.StatusInternalServerError, "open failed: "+msg)
			return
		}
		c.String(http.StatusOK, "opened")
	}
}

func secureCleanAbs(p string) (string, error) {
	if p == "" {
		return "", errors.New("empty")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func underAnyRoot(p string, roots []string) bool {
	for _, r := range roots {
		if strings.HasPrefix(p, r+"/") || p == r {
			return true
		}
	}
	return false
}
