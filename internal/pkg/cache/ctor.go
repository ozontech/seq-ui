package cache

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

var caches map[string]Cache

func Init(ctx context.Context, cfg *config.Cache) error {
	caches = make(map[string]Cache, len(cfg.Redis)+len(cfg.Inmemory))
	withInmemIDs := make(map[string]struct{})

	for _, redis := range cfg.Redis {
		var inmem *config.Inmemory
		if redis.WithInmemID != "" {
			inmem = cfg.InmemByID(redis.WithInmemID)
			withInmemIDs[redis.WithInmemID] = struct{}{}
		}

		c, err := newRedisOrInmemWithRedis(ctx, &redis, inmem)
		if err != nil {
			return fmt.Errorf("init redis cache %q: %w", redis.ID, err)
		}
		caches[redis.ID] = c
	}

	for _, inmem := range cfg.Inmemory {
		if _, ok := withInmemIDs[inmem.ID]; ok {
			continue
		}

		c, err := newInmemoryCache(inmem)
		if err != nil {
			return fmt.Errorf("init inmemory cache %q: %w", inmem.ID, err)
		}
		caches[inmem.ID] = c
	}

	return nil
}

func ByID(id string) Cache {
	return caches[id]
}

func newRedisOrInmemWithRedis(ctx context.Context, redisCfg *config.Redis, inmemCfg *config.Inmemory) (Cache, error) {
	redis, err := newRedisCache(ctx, redisCfg)
	if err != nil {
		return nil, err
	}
	if inmemCfg == nil {
		return redis, nil
	}

	inmem, err := newInmemoryCache(*inmemCfg)
	if err != nil {
		return nil, fmt.Errorf("init inmemory cache %q: %w", inmemCfg.ID, err)
	}

	return &inmemWithRedis{
		inmem: inmem,
		redis: redis,
	}, nil
}
