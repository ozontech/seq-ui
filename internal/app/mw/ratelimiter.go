package mw

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
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
	"AdminService":       {},
}

type rateLimiter interface {
	RateLimit(string) (bool, ratelimiter.RateLimitInfo, error)
}

type RateLimiter struct {
	rateLimiter
	perHandler bool
}

// RateLimitersByEnv contains rate limiters by env->api->user, empty env is root options.
type RateLimitersByEnv map[string]map[string]map[string]RateLimiter

func (ratelimiters RateLimitersByEnv) get(ctx context.Context) map[string]map[string]RateLimiter {
	if rl, ok := ratelimiters[types.GetEnvKey(ctx)]; ok {
		return rl
	}
	return ratelimiters[""]
}

func NewRateLimiters(apiToRateLimiters config.ApiToRateLimiters) (map[string]map[string]RateLimiter, error) {
	result := make(map[string]map[string]RateLimiter, len(apiToRateLimiters))

	for apiName, rateLimiters := range apiToRateLimiters {
		result[apiName] = make(map[string]RateLimiter)

		defaultLimiter, err := newRateLimiter(apiName, rateLimiters.Default)
		if err != nil {
			return nil, fmt.Errorf("init default rate limiter: %w", err)
		}
		result[apiName][RateLimiterDefaultUser] = defaultLimiter

		for user, userRateLimiter := range rateLimiters.SpecialUsers {
			limiter, err := newRateLimiter(apiName, userRateLimiter)
			if err != nil {
				return nil, fmt.Errorf("init user %q rate limiter: %w", user, err)
			}
			result[apiName][user] = limiter
		}
	}

	return result, nil
}

func newRateLimiter(api string, cfg config.RateLimiter) (RateLimiter, error) {
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
