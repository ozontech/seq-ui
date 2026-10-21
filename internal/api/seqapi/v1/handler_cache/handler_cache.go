package handlercache

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
)

type Cache struct {
	Cache     cache.Cache
	KeyPrefix string
	TTL       time.Duration
}

type Caches struct {
	Events       *Cache
	Fields       *Cache
	LogsLifespan *Cache
}

func New(env string, cfg config.SeqAPICaches) Caches {
	keyPrefix := "seqapi_"
	if env != "" {
		keyPrefix += env + "_"
	}

	return Caches{
		Events:       newCache(cfg.Events, keyPrefix+"event_"),
		Fields:       newCache(cfg.Fields, keyPrefix+"fields"),
		LogsLifespan: newCache(cfg.LogsLifespan, keyPrefix+"logs_lifespan"),
	}
}

func newCache(cfg *config.HandlerCache, keyPrefix string) *Cache {
	if cfg == nil {
		return nil
	}

	c, ok := cache.ByID(cfg.ID)
	if !ok {
		return nil
	}

	return &Cache{Cache: c, KeyPrefix: keyPrefix, TTL: cfg.TTL}
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	if c == nil {
		return "", false
	}

	value, err := c.Cache.Get(ctx, c.KeyPrefix+key)
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			logger.Error("failed to get value from cache", zap.String("key", c.KeyPrefix+key), zap.Error(err))
		}
		return "", false
	}

	return value, true
}

func (c *Cache) Set(ctx context.Context, key, value string) {
	if c == nil {
		return
	}

	if err := c.Cache.SetWithTTL(ctx, c.KeyPrefix+key, value, c.TTL); err != nil {
		logger.Error("failed to set value to cache", zap.String("key", c.KeyPrefix+key), zap.Error(err))
	}
}
