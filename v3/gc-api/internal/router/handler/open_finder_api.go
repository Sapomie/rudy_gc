package handler

import (
	"context"
	"html"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type openFinderReq struct {
	Path string `json:"path"`
}

func (h *MovieDetailAPI) OpenFinder(allowRoots []string) gin.HandlerFunc {
	normalizedRoots := make([]string, 0, len(allowRoots))
	for _, p := range allowRoots {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		normalizedRoots = append(normalizedRoots, filepath.Clean(abs))
	}

	return func(c *gin.Context) {
		if runtime.GOOS != "darwin" {
			c.String(http.StatusNotImplemented, "open finder only supports darwin")
			return
		}

		var req openFinderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, "bad json")
			return
		}
		if strings.TrimSpace(req.Path) == "" {
			c.String(http.StatusBadRequest, "missing path")
			return
		}

		target, err := cleanAbsPath(req.Path)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid path")
			return
		}
		if !pathUnderAnyRoot(target, normalizedRoots) {
			c.String(http.StatusForbidden, "path not allowed: "+html.EscapeString(target))
			return
		}

		fi, err := os.Stat(target)
		if err != nil {
			c.String(http.StatusNotFound, "not found: "+html.EscapeString(err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var cmd *exec.Cmd
		if fi.IsDir() {
			cmd = exec.CommandContext(ctx, "open", target)
		} else {
			cmd = exec.CommandContext(ctx, "open", "-R", target)
		}
		out, err := cmd.CombinedOutput()
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

func cleanAbsPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func pathUnderAnyRoot(p string, roots []string) bool {
	for _, root := range roots {
		if p == root || strings.HasPrefix(p, root+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
