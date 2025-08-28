package cache

import (
	"context"
	"errors"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value string) error
	SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Del(ctx context.Context, key string)
}

var ErrNotFound = errors.New("not found")
