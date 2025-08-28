package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/redisclient"
	"github.com/ozontech/seq-ui/metric"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func newRedisCache(ctx context.Context, cfg *config.Redis) (*redisCache, error) {
	client, err := redisclient.New(ctx, cfg)
	if err != nil {
		metric.ServerCacheRedisMisses.Inc()
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	metric.ServerCacheRedisHits.Inc()
	return &redisCache{client: client}, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	return c.SetWithTTL(ctx, key, value, 0)
}

func (c *redisCache) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if err == nil {
		return value, nil
	}

	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}

	return "", err
}

func (c *redisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	value, err := c.client.TTL(ctx, key).Result()
	if err == nil {
		return value, nil
	}

	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}

	return 0, err
}

func (c *redisCache) Del(ctx context.Context, key string) {
	c.client.Del(ctx, key)
}
