package cache

import (
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/metric"
)

const defaultCost = 1

type inmemoryCache struct {
	cache *ristretto.Cache[string, string]
}

func newInmemoryCache(cfg config.InmemoryCache) (*inmemoryCache, error) {
	c, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
		Metrics:     false,
	})
	if err != nil {
		return nil, err
	}

	return &inmemoryCache{cache: c}, nil
}

func (c *inmemoryCache) Set(ctx context.Context, key string, value string) error {
	return c.SetWithTTL(ctx, key, value, 0)
}

func (c *inmemoryCache) SetWithTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	if !c.cache.SetWithTTL(key, value, defaultCost, ttl) {
		return errors.New("failed to add the key-value item to the cache")
	}

	return nil
}

func (c *inmemoryCache) Get(_ context.Context, key string) (string, error) {
	val, ok := c.cache.Get(key)
	if !ok {
		metric.ServerCacheInmemoryMisses.Inc()
		return "", ErrNotFound
	}

	metric.ServerCacheInmemoryHits.Inc()
	return val, nil
}

func (c *inmemoryCache) GetTTL(_ context.Context, key string) (time.Duration, error) {
	val, ok := c.cache.GetTTL(key)
	if !ok {
		return val, ErrNotFound
	}

	return val, nil
}

func (c *inmemoryCache) Del(_ context.Context, key string) {
	c.cache.Del(key)
}
