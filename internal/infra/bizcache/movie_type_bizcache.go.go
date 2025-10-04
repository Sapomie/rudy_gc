// internal/infra/bizcache/movie_type_bizcache.go
package bizcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type MovieTypeBizCache struct {
	rdb     *redis.Redis
	prefix  string
	version string
	ttl     time.Duration
}

func NewMovieTypeBizCache(rdb *redis.Redis, prefix, version string, ttl time.Duration) movie_repo.MovieTypeCache {
	return &MovieTypeBizCache{rdb: rdb, prefix: prefix, version: version, ttl: ttl}
}

func (c *MovieTypeBizCache) key(javId string) string {
	return fmt.Sprintf("%s:%s:%s", c.prefix, c.version, javId)
}

func (c *MovieTypeBizCache) GetMovieType(ctx context.Context, javId string) (*types.MovieType, error) {
	val, err := c.rdb.Get(c.key(javId))
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

func (c *MovieTypeBizCache) SetMovieType(ctx context.Context, javId string, v *types.MovieType, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.ttl
	}
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Setex(c.key(javId), string(bs), int(ttl/time.Second))
}

func (c *MovieTypeBizCache) DelMovieType(ctx context.Context, javId string) error {
	_, err := c.rdb.Del(c.key(javId))
	return err
}
