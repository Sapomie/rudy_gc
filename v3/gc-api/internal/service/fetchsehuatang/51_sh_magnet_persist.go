package fetchsehuatang

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) persistSHMagnets(ctx context.Context, thread *moviex.ShSehuatangThread, resources []parsedMagnet, seenAt int64) error {
	if len(resources) == 0 {
		return nil
	}
	if s == nil || s.deps == nil || s.deps.ShMagnetModel == nil {
		return fmt.Errorf("sh magnet model is nil")
	}
	for _, row := range buildSHMagnetRows(thread, resources, seenAt) {
		if err := s.deps.ShMagnetModel.Upsert(ctx, row); err != nil {
			return fmt.Errorf("写入 sh_magnet 失败: info_hash=%s: %w", row.InfoHash, err)
		}
	}
	return nil
}

func buildSHMagnetRows(thread *moviex.ShSehuatangThread, resources []parsedMagnet, seenAt int64) []*moviex.ShMagnet {
	if thread == nil || len(resources) == 0 {
		return nil
	}
	if seenAt <= 0 {
		seenAt = time.Now().Unix()
	}
	seen := make(map[string]struct{}, len(resources))
	out := make([]*moviex.ShMagnet, 0, len(resources))
	for _, resource := range resources {
		infoHash := strings.ToUpper(strings.TrimSpace(resource.InfoHash))
		if infoHash == "" {
			continue
		}
		if _, ok := seen[infoHash]; ok {
			continue
		}
		seen[infoHash] = struct{}{}
		out = append(out, &moviex.ShMagnet{
			ThreadTID:     strings.TrimSpace(thread.ThreadTID),
			MovieJavId:    strings.TrimSpace(thread.MovieJavId),
			MovieName:     strings.TrimSpace(thread.MovieName),
			MovieType:     strings.TrimSpace(thread.MovieType),
			SourceForumID: strings.TrimSpace(thread.SourceForumID),
			PostTime:      thread.PostTime,
			MagnetURL:     strings.TrimSpace(resource.MagnetURL),
			InfoHash:      infoHash,
			CreatedOn:     seenAt,
			UpdatedOn:     seenAt,
		})
	}
	return out
}
