// internal/domain/movie/mt__movie_type.go
package movie

import (
	"context"
	"time"

	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const movieTypeTTL = 7 * 24 * time.Hour // 也可放 config

func (s *Service) GetMovieType(ctx context.Context, javId string) (*types.MovieType, error) {
	// 1) 先查缓存
	if s.deps.MovieTypeCache != nil {
		if v, err := s.deps.MovieTypeCache.GetMovieType(ctx, javId); err == nil && v != nil {
			return v, nil
		}
	}

	// 2) 回源：用你现有 repo 聚合拼装 MovieType（略）
	mt, err := s.buildMovieTypeFromRepos(ctx, javId)
	if err != nil || mt == nil {
		return mt, err
	}

	// 3) 写缓存（忽略错误不影响主流程）
	if s.deps.MovieTypeCache != nil {
		_ = s.deps.MovieTypeCache.SetMovieType(ctx, javId, mt, movieTypeTTL)
	}

	return mt, nil
}

// 单个失效
func (s *Service) InvalidateMovieType(ctx context.Context, javId string) {
	if s.deps.MovieTypeCache == nil || javId == "" {
		return
	}
	if err := s.deps.MovieTypeCache.DelMovieType(ctx, javId); err != nil {
		logx.WithContext(ctx).Errorf("del MovieType cache failed, javId=%s, err=%v", javId, err)
	} else {
		logx.WithContext(ctx).Infof("del MovieType cache ok, javId=%s", javId)
	}
}

// InvalidateMovieTypes 精确失效：多个影片（去重 + 跳过空值）
func (s *Service) InvalidateMovieTypes(ctx context.Context, javIds ...string) {
	if s.deps.MovieTypeCache == nil || len(javIds) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(javIds))
	for _, id := range javIds {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		if err := s.deps.MovieTypeCache.DelMovieType(ctx, id); err != nil {
			logx.WithContext(ctx).Errorf("del MovieType cache failed, javId=%s, err=%v", id, err)
		} else {
			logx.WithContext(ctx).Infof("del MovieType cache ok, javId=%s", id)
		}
	}
}
