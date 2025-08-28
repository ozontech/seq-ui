package cache

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/logger"
	"go.uber.org/zap"
)

func NewInmemoryWithRedisOrInmemory(ctx context.Context, cfg config.Cache) (Cache, error) {
	return newRedisBasedOrInmemory(ctx, cfg, true)
}

func NewRedisOrInmemory(ctx context.Context, cfg config.Cache) (Cache, error) {
	return newRedisBasedOrInmemory(ctx, cfg, false)
}

func newRedisBasedOrInmemory(ctx context.Context, cfg config.Cache, withInmem bool) (Cache, error) {
	inmem, err := newInmemoryCache(cfg.Inmemory)
	if err != nil {
		return nil, fmt.Errorf("init inmemory cache: %w", err)
	}

	if cfg.Redis == nil {
		logger.Warn("redis cache config is nil; inmemory cache will be used instead")
		return inmem, nil
	}

	redis, err := newRedisCache(ctx, cfg.Redis)
	if err != nil {
		logger.Warn("failed to init redis cache; inmemory cache will be used instead", zap.Error(err))
		return inmem, nil
	}

	if withInmem {
		return &inmemWithRedis{
			inmem: inmem,
			redis: redis,
		}, nil
	}

	return redis, nil
}
