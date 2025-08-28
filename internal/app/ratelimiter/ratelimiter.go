package ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

const rateLimiterQuantityPerCall = 1

type RateLimitInfo struct {
	Limit      int
	Remaining  int
	ResetAfter time.Duration
	RetryAfter time.Duration
}

type RateLimiter struct {
	gcraRL *throttled.GCRARateLimiterCtx
}

func New(cfg config.RateLimiter) (*RateLimiter, error) {
	if cfg.RatePerSec <= 0 {
		return nil, errors.New(
			"invalid rate limiter config: rate_per_sec must be greater than zero",
		)
	}
	if cfg.MaxBurst < 0 {
		return nil, errors.New(
			"invalid rate limiter config: max_burst must be non-negative",
		)
	}
	rlStore, err := memstore.NewCtx(cfg.StoreMaxKeys)
	if err != nil {
		return nil, err
	}
	rlQuota := throttled.RateQuota{
		MaxRate:  throttled.PerSec(cfg.RatePerSec),
		MaxBurst: cfg.MaxBurst,
	}
	gcraRL, err := throttled.NewGCRARateLimiterCtx(rlStore, rlQuota)
	if err != nil {
		return nil, err
	}
	return &RateLimiter{
		gcraRL: gcraRL,
	}, nil
}

func (rl *RateLimiter) RateLimit(key string) (bool, RateLimitInfo, error) {
	limit, rlc, err := rl.gcraRL.RateLimitCtx(context.Background(), key, rateLimiterQuantityPerCall)
	rlInfo := RateLimitInfo{
		Limit:      rlc.Limit,
		Remaining:  rlc.Remaining,
		ResetAfter: rlc.ResetAfter,
		RetryAfter: rlc.RetryAfter,
	}
	return limit, rlInfo, err
}
