package migrate

import (
	"context"
	"log"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"
)

func (s *Service) MigrateLocalCover() error {
	ctx := context.Background()
	items, err := s.deps.ItemRepo.FindByDownloadCoverStatus(ctx, consts.ItemCoverNone)
	if err != nil {
		return err
	}

	var count int
	for _, item := range items {
		murl, err := s.xModel.MurlModel.FindOneByJavId(ctx, item.JavId)
		if err != nil {
			return err
		}
		p := types.MurlPatch{
			JacketImgLocal: &murl.JacketImgLocal,
		}
		err = s.deps.MurlRepo.UpdatePartialByJavId(ctx, item.JavId, p)
		if err != nil {
			return err
		}
		s.movieSvc.InvalidateMovieType(ctx, item.JavId)

		err = s.deps.ItemRepo.UpdatePartialByJavId(ctx, item.JavId, types.ItemPatch{
			HasDownloadCover: ptr.Int64(consts.ItemCoverOK),
		})
		if err != nil {
			return err
		}

		count++
		if count%100 == 0 {
			log.Printf("migrate old cover count: %d/%d", count, len(items))
		}
	}

	return nil
}
