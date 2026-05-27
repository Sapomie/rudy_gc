package dep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy-gc-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type MovieTypeCache interface {
	GetMovieCard(ctx context.Context, javID string) (*types.MovieCard, error)
	SetMovieCard(ctx context.Context, javID string, card *types.MovieCard, ttl time.Duration) error
	DelMovieCard(ctx context.Context, javID string) error
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

func (c *movieTypeBizCache) key(javID string) string {
	return fmt.Sprintf("%s:%s:%s", c.prefix, c.version, javID)
}

func (c *movieTypeBizCache) GetMovieCard(ctx context.Context, javID string) (*types.MovieCard, error) {
	if c == nil || c.rdb == nil || javID == "" {
		return nil, nil
	}
	val, err := c.rdb.GetCtx(ctx, c.key(javID))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if strings.TrimSpace(val) == "" {
		_ = c.DelMovieCard(ctx, javID)
		return nil, nil
	}
	var card types.MovieCard
	if err := json.Unmarshal([]byte(val), &card); err != nil {
		_ = c.DelMovieCard(ctx, javID)
		return nil, nil
	}
	return &card, nil
}

func (c *movieTypeBizCache) SetMovieCard(ctx context.Context, javID string, card *types.MovieCard, ttl time.Duration) error {
	if c == nil || c.rdb == nil || javID == "" || card == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = c.ttl
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	bs, err := json.Marshal(card)
	if err != nil {
		return err
	}
	return c.rdb.SetexCtx(ctx, c.key(javID), string(bs), int(ttl/time.Second))
}

func (c *movieTypeBizCache) DelMovieCard(ctx context.Context, javID string) error {
	if c == nil || c.rdb == nil || javID == "" {
		return nil
	}
	_, err := c.rdb.DelCtx(ctx, c.key(javID))
	return err
}
