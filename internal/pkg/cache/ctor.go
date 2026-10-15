package cache

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

var caches map[string]Cache

func Init(ctx context.Context, cfg *config.Cache) error {
	caches = make(map[string]Cache, len(cfg.Redis)+len(cfg.Inmemory))

	for i := range cfg.Inmemory {
		inmemCfg := &cfg.Inmemory[i]
		inmem, err := newInmemoryCache(inmemCfg)
		if err != nil {
			return fmt.Errorf("init inmemory cache %q: %w", inmemCfg.ID, err)
		}
		caches[inmemCfg.ID] = inmem
	}

	for i := range cfg.Redis {
		redisCfg := &cfg.Redis[i]

		var inmem Cache
		if redisCfg.WithInmemID != "" {
			inmem = caches[redisCfg.WithInmemID]
		}

		redis, err := newRedisCache(ctx, redisCfg, inmem)
		if err != nil {
			return fmt.Errorf("init redis cache %q: %w", redisCfg.ID, err)
		}
		caches[redisCfg.ID] = redis
	}

	return nil
}

// ByID returns cache by id and true, false if cache is not found.
func ByID(id string) (Cache, bool) {
	cache, ok := caches[id]
	return cache, ok
}
