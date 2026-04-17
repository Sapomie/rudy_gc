package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) ProbeMediaMetaByID(ctx context.Context, id int64) (*types.MediaProbeMetaResult, error) {
	row, err := s.deps.WMediaModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil || row.SourceType != consts.WMediaSourceNative {
		return nil, types.ErrNotFound
	}

	fullPath := buildMediaAbsolutePath(row)
	if strings.TrimSpace(fullPath) == "" {
		return nil, fmt.Errorf("媒体路径为空")
	}
	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("媒体文件不存在: %w", err)
	}

	meta, err := probeVideoMeta(fullPath)
	if err != nil {
		return nil, err
	}

	updated := *row
	updated.Width = meta.width
	updated.Height = meta.height
	updated.BitRate = meta.bitRate
	updated.Duration = meta.duration
	updated.FrameAverage = meta.frameAverage
	updated.NeedScanMeta = consts.FilmMetaDataNoNeedScan
	updated.UpdatedOn = time.Now().Unix()

	if err := s.deps.WMediaModel.Update(ctx, &updated); err != nil {
		return nil, err
	}
	s.invalidateMovieTypeCaches(ctx, updated.MovieJavId)

	return &types.MediaProbeMetaResult{
		Id:              updated.Id,
		Height:          updated.Height,
		DurationMinutes: formatMediaDurationMinutes(updated.Duration),
		BitRate:         updated.BitRate,
		FrameAverage:    formatMediaFrameAverage(updated.FrameAverage),
	}, nil
}

func buildMediaAbsolutePath(row *moviex.WMedia) string {
	if row == nil {
		return ""
	}
	fileName := strings.TrimSpace(row.FileName)
	if fileName == "" {
		return ""
	}

	fullDir := strings.TrimSpace(row.FullDir)
	if fullDir != "" {
		return filepath.Join(fullDir, fileName)
	}

	rootDir := strings.TrimSpace(row.RootDir)
	if rootDir != "" {
		return filepath.Join(rootDir, fileName)
	}

	return fileName
}
