package movie

import (
	"context"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"

	"github.com/pkg/errors"
)

func (s *MovieService) AddToDownloadLater(ctx context.Context, javId string) (int64, error) {
	patch := types.MinfoPatch{
		NeedDownload: ptr.Int64(consts.MovieNeedDownLoadOK),
	}
	err := s.deps.MinfoRepo.UpdatePartialByJavId(ctx, javId, patch)
	if err != nil {
		return 0, errors.Wrap(err, "add to download later failed")
	}
	s.InvalidateMovieType(ctx, javId)
	minfo, err := s.deps.MinfoRepo.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, errors.Wrap(err, "find to download later failed")
	}

	return minfo.NeedDownload, nil
}

func (s *MovieService) RemoveFromDownloadLater(ctx context.Context, javId string) (int64, error) {
	patch := types.MinfoPatch{
		NeedDownload: ptr.Int64(1),
	}
	if err := s.deps.MinfoRepo.UpdatePartialByJavId(ctx, javId, patch); err != nil {
		return 0, errors.Wrap(err, "remove from download later failed")
	}

	s.InvalidateMovieType(ctx, javId)

	minfo, err := s.deps.MinfoRepo.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, errors.Wrap(err, "find after remove download later failed")
	}

	return minfo.NeedDownload, nil
}
