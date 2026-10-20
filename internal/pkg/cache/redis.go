package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/redisclient"
	"github.com/ozontech/seq-ui/metric"
)

const keysBatchSize = 10000

type redisCache struct {
	client *redis.Client
	prefix string
}

func newRedisCache(ctx context.Context, cfg *config.Redis, inmem Cache) (Cache, error) {
	client, err := redisclient.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	rc := &redisCache{client: client, prefix: cfg.KeyPrefix}
	if inmem == nil {
		return rc, nil
	}

	return &redisWithInmem{
		redis: rc,
		inmem: inmem,
	}, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	return c.SetWithTTL(ctx, key, value, 0)
}

func (c *redisCache) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, c.withPrefix(key), value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, c.withPrefix(key)).Result()
	if err == nil {
		metric.ServerCacheRedisHits.Inc()
		return value, nil
	}

	if errors.Is(err, redis.Nil) {
		metric.ServerCacheRedisMisses.Inc()
		return "", ErrNotFound
	}

	return "", err
}

func (c *redisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	value, err := c.client.TTL(ctx, c.withPrefix(key)).Result()
	if err == nil {
		return value, nil
	}

	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}

	return 0, err
}

func (c *redisCache) Del(ctx context.Context, key string) {
	c.client.Del(ctx, c.withPrefix(key))
}

func (c *redisCache) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.withPrefix(key), value, ttl).Result()
}

func (c *redisCache) KeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	var (
		cursor          uint64
		result, partRes []string
		err             error
	)

	for {
		partRes, cursor, err = c.client.Scan(ctx, cursor, c.withPrefix(pattern), keysBatchSize).Result()
		if err != nil {
			return nil, fmt.Errorf("redis scan: %w", err)
		}

		for _, key := range partRes {
			result = append(result, strings.TrimPrefix(key, c.prefix))
		}

		if cursor == 0 {
			break
		}
	}

	return result, nil
}

func (c *redisCache) MGet(ctx context.Context, keys ...string) ([]any, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.withPrefix(key)
	}

	return c.client.MGet(ctx, prefixedKeys...).Result()
}

func (c *redisCache) withPrefix(key string) string {
	return c.prefix + key
}
