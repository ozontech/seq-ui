package cache

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	config "github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/logger"
)

func FromConfig(ctx context.Context, cfg *config.Cache) (map[string]Cache, error) {
	caches := make(map[string]Cache, len(cfg.Redis)+len(cfg.Inmemory))

	for i := range cfg.Redis {
		redis := &cfg.Redis[i]
		var inmem *config.InmemoryCache
		if redis.WithInmemID != "" {
			inmem = cfg.InmemByID(redis.WithInmemID)
		}

		cache, err := New(ctx, inmem, redis)
		if err != nil {
			return nil, fmt.Errorf("new cache %q: %w", redis.ID, err)
		}

		caches[redis.ID] = cache
	}

	for i := range cfg.Inmemory {
		inmem := &cfg.Inmemory[i]
		if hasWithInmem(inmem.ID, cfg.Redis) {
			continue
		}

		cache, err := newInmemoryCache(*inmem)
		if err != nil {
			return nil, fmt.Errorf("new inmemory cache %q: %w", inmem.ID, err)
		}

		caches[inmem.ID] = cache
	}

	return caches, nil
}

func New(ctx context.Context, inmemCfg *config.InmemoryCache, redisCfg *config.Redis) (Cache, error) {
	redis, redisErr := newRedisCache(ctx, redisCfg)
	if inmemCfg == nil {
		if redisErr != nil {
			return nil, fmt.Errorf("init redis cache: %w", redisErr)
		}
		return redis, nil
	}

	inmem, err := newInmemoryCache(*inmemCfg)
	if err != nil {
		return nil, fmt.Errorf("init inmemory cache: %w", err)
	}

	if redisErr != nil {
		logger.Warn("failed to init redis cache; inmemory cache will be used instead", zap.Error(err))
		return inmem, err
	}

	return &inmemWithRedis{
		inmem: inmem,
		redis: redis,
	}, nil
}

func hasWithInmem(inmemID string, redisCfgs []config.Redis) bool {
	for i := range redisCfgs {
		if redisCfgs[i].WithInmemID == inmemID {
			return true
		}
	}

	return false
}
