package svc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type MovieTypeCache interface {
	GetMovieType(ctx context.Context, javId string) (*types.MovieType, error)
	SetMovieType(ctx context.Context, javId string, v *types.MovieType, ttl time.Duration) error
	DelMovieType(ctx context.Context, javId string) error
}

type movieTypeBizCache struct {
	rdb     *redis.Redis
	prefix  string
	version string
	ttl     time.Duration
}

func newMovieTypeBizCache(rdb *redis.Redis, prefix, version string, ttl time.Duration) MovieTypeCache {
	return &movieTypeBizCache{rdb: rdb, prefix: prefix, version: version, ttl: ttl}
}

func (c *movieTypeBizCache) key(javId string) string {
	return fmt.Sprintf("%s:%s:%s", c.prefix, c.version, javId)
}

func (c *movieTypeBizCache) GetMovieType(ctx context.Context, javId string) (*types.MovieType, error) {
	val, err := c.rdb.GetCtx(ctx, c.key(javId))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var mt types.MovieType
	if err := json.Unmarshal([]byte(val), &mt); err != nil {
		return nil, err
	}
	return &mt, nil
}

func (c *movieTypeBizCache) SetMovieType(ctx context.Context, javId string, v *types.MovieType, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.ttl
	}
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.SetexCtx(ctx, c.key(javId), string(bs), int(ttl/time.Second))
}

func (c *movieTypeBizCache) DelMovieType(ctx context.Context, javId string) error {
	_, err := c.rdb.DelCtx(ctx, c.key(javId))
	return err
}
