package mw

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/ratelimiter"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/swagger"
)

const RateLimiterDefaultUser = "_"

var grpcServices = map[string]struct{}{
	"SeqAPIService":      {},
	"UserProfileService": {},
	"DashboardsService":  {},
	"MassExportService":  {},
	"ErrorGroupsService": {},
}

type rateLimiter interface {
	RateLimit(string) (bool, ratelimiter.RateLimitInfo, error)
}

type RateLimiter struct {
	rateLimiter
	perHandler bool
}

func NewRateLimiter(api string, cfg config.RateLimiter) (RateLimiter, error) {
	if _, ok := grpcServices[api]; !ok {
		spec := swagger.GetSpecEx()
		if !spec.HasApi(api) {
			return RateLimiter{}, fmt.Errorf("invalid rate limiter api %q", api)
		}
	}

	limiter, err := ratelimiter.New(cfg)
	if err != nil {
		return RateLimiter{}, fmt.Errorf("init %q rate limiter: %w", api, err)
	}

	return RateLimiter{
		rateLimiter: limiter,
		perHandler:  cfg.PerHandler,
	}, nil
}

func handleUserRateLimit(
	ctx context.Context, userToRateLimiter map[string]RateLimiter, handler string,
) (bool, ratelimiter.RateLimitInfo, error) {
	userName, err := types.GetUserKey(ctx)
	if err != nil {
		userName = RateLimiterDefaultUser
	}

	limiter, ok := userToRateLimiter[userName]
	if !ok {
		limiter = userToRateLimiter[RateLimiterDefaultUser]
	}

	key := userName
	if limiter.perHandler {
		key = fmt.Sprintf("%s_%s", userName, handler)
	}

	limit, rlc, err := limiter.RateLimit(key)
	if limit {
		metric.ServerRateLimits.Inc()
	}

	return limit, rlc, err
}

func writeRateLimitHTTPHeaders(w http.ResponseWriter, rlInfo ratelimiter.RateLimitInfo) {
	if v := rlInfo.Limit; v >= 0 {
		w.Header().Add("X-RateLimit-Limit", strconv.Itoa(v))
	}

	if v := rlInfo.Remaining; v >= 0 {
		w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(v))
	}

	if v := rlInfo.ResetAfter; v >= 0 {
		vi := int(math.Ceil(v.Seconds()))
		w.Header().Add("X-RateLimit-Reset", strconv.Itoa(vi))
	}

	if v := rlInfo.RetryAfter; v >= 0 {
		vi := int(math.Ceil(v.Seconds()))
		w.Header().Add("Retry-After", strconv.Itoa(vi))
	}
}
