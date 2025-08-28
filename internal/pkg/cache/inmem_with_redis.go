package cache

import (
	"context"
	"errors"
	"time"

	"github.com/ozontech/seq-ui/logger"
	"go.uber.org/zap"
)

type inmemWithRedis struct {
	inmem *inmemoryCache
	redis *redisCache
}

func (c *inmemWithRedis) Set(ctx context.Context, key string, value string) error {
	return c.SetWithTTL(ctx, key, value, 0)
}

func (c *inmemWithRedis) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := c.redis.SetWithTTL(ctx, key, value, ttl); err != nil {
		logger.Error("failed to set value in redis sub cache", zap.String("key", key), zap.Error(err))
	}

	return c.inmem.SetWithTTL(ctx, key, value, ttl)
}

func (c *inmemWithRedis) Get(ctx context.Context, key string) (string, error) {
	val, err := c.inmem.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	val, err = c.redis.Get(ctx, key)
	if err == nil {
		if ttl, err := c.GetTTL(ctx, key); err == nil {
			_ = c.inmem.SetWithTTL(ctx, key, val, ttl)
		}

		return val, nil
	}

	if !errors.Is(err, ErrNotFound) {
		logger.Error("failed to get value from redis sub cache", zap.String("key", key), zap.Error(err))
	}

	return "", err
}

func (c *inmemWithRedis) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	val, err := c.inmem.GetTTL(ctx, key)
	if err == nil {
		return val, nil
	}

	val, err = c.redis.GetTTL(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrNotFound) {
		logger.Error("failed to get ttl from redis sub cache", zap.String("key", key), zap.Error(err))
	}

	return 0, err
}

func (c *inmemWithRedis) Del(ctx context.Context, key string) {
	c.redis.Del(ctx, key)
	c.inmem.Del(ctx, key)
}
