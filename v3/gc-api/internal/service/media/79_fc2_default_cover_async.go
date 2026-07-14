package media

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/threading"
)

var fc2DefaultCoverRunning sync.Map

func (s *Service) EnqueueFC2DefaultCovers(mediaIDs []int64) {
	cleanIDs := uniqueFC2MediaIDs(mediaIDs)
	if len(cleanIDs) == 0 {
		return
	}
	threading.GoSafe(func() {
		s.runFC2DefaultCoverIDs(context.Background(), cleanIDs)
	})
}

func (s *Service) EnqueueFC2DefaultCoverBackfill() {
	threading.GoSafe(func() {
		if s == nil || s.deps == nil || s.deps.FC2MediaModel == nil {
			return
		}
		rows, err := s.deps.FC2MediaModel.ListAll(context.Background())
		if err != nil {
			s.deps.Log.Errorf("FC2默认联系表封面补齐查询失败: %v", err)
			return
		}
		mediaIDs := make([]int64, 0, len(rows))
		for _, row := range rows {
			if row != nil && row.ID > 0 {
				mediaIDs = append(mediaIDs, row.ID)
			}
		}
		s.runFC2DefaultCoverIDs(context.Background(), mediaIDs)
	})
}

func (s *Service) runFC2DefaultCoverIDs(ctx context.Context, mediaIDs []int64) {
	for _, mediaID := range uniqueFC2MediaIDs(mediaIDs) {
		if _, loaded := fc2DefaultCoverRunning.LoadOrStore(mediaID, struct{}{}); loaded {
			continue
		}
		err := func() error {
			defer fc2DefaultCoverRunning.Delete(mediaID)
			return s.ensureFC2DefaultCover(ctx, mediaID)
		}()
		if err != nil {
			s.deps.Log.Errorf("FC2默认联系表封面生成失败: media_id=%d err=%v", mediaID, err)
		}
	}
}

func (s *Service) ensureFC2DefaultCover(ctx context.Context, mediaID int64) error {
	if s == nil || s.deps == nil || s.deps.FC2MediaModel == nil || s.deps.FC2MovieModel == nil || s.deps.FC2MediaImageModel == nil {
		return fmt.Errorf("FC2默认联系表封面服务依赖为空")
	}
	mediaRow, err := s.deps.FC2MediaModel.FindOne(ctx, mediaID)
	if err != nil {
		return err
	}
	movieRow, err := s.deps.FC2MovieModel.FindOne(ctx, mediaRow.FC2MovieID)
	if err != nil {
		return err
	}
	imageRows, err := s.deps.FC2MediaImageModel.ListByMediaID(ctx, mediaRow.ID)
	if err != nil {
		return err
	}
	if coverImage := ResolveFC2CoverImage(mediaRow.ID, movieRow.CoverImageID, imageRows); coverImage != nil {
		if movieRow.CoverImageID == coverImage.ID {
			return nil
		}
		return s.saveFC2CoverImage(ctx, mediaRow.FC2MovieID, coverImage.ID)
	}
	_, err = s.GenerateFC2DefaultCover(ctx, mediaRow)
	return err
}

func uniqueFC2MediaIDs(mediaIDs []int64) []int64 {
	seen := make(map[int64]struct{}, len(mediaIDs))
	out := make([]int64, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		if mediaID <= 0 {
			continue
		}
		if _, ok := seen[mediaID]; ok {
			continue
		}
		seen[mediaID] = struct{}{}
		out = append(out, mediaID)
	}
	return out
}
