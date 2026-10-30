package cache

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

var caches map[string]Cache

func Init(ctx context.Context, cfg *config.Cache) error {
	caches = make(map[string]Cache, len(cfg.Redis)+len(cfg.Inmemory))

	for id, inmemCfg := range cfg.Inmemory {
		inmem, err := newInmemoryCache(inmemCfg)
		if err != nil {
			return fmt.Errorf("init inmemory cache %q: %w", id, err)
		}
		caches[id] = inmem
	}

	for id, redisCfg := range cfg.Redis {
		var inmem Cache
		if redisCfg.WithInmemID != "" {
			inmem = caches[redisCfg.WithInmemID]
		}

		redis, err := newRedisCache(ctx, redisCfg, inmem)
		if err != nil {
			return fmt.Errorf("init redis cache %q: %w", id, err)
		}
		caches[id] = redis
	}

	return nil
}

// ByID returns cache by id and true, false if cache is not found.
func ByID(id string) (Cache, bool) {
	cache, ok := caches[id]
	return cache, ok
}
