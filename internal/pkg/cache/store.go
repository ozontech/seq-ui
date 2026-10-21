package cache

import (
	"context"
	"time"
)

type Store interface {
	Cache
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	KeysByPattern(ctx context.Context, pattern string) ([]string, error)
	MGet(ctx context.Context, keys ...string) ([]any, error)
}

func StoreByID(id string) (Store, bool) {
	store, ok := caches[id].(Store)
	return store, ok
}
