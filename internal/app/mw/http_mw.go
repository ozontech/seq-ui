package mw

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/swagger"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type routePatternKey struct{}

func setRoutePattern(ctx context.Context, pattern string) context.Context {
	l := len(pattern)
	if l > 1 && pattern[l-1] == '/' {
		pattern = pattern[:l-1]
	}
	return context.WithValue(ctx, routePatternKey{}, pattern)
}

func getRoutePattern(ctx context.Context) string {
	if v := ctx.Value(routePatternKey{}); v != nil {
		return v.(string)
	}
	return ""
}

func HTTPLogInterceptor(logger *tracing.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error(r.Context(), "failed to read request body", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

			ww := httputil.NewWriter(w)

			fullMethod := strings.Join(
				[]string{r.Method, r.RequestURI, r.Proto}, " ",
			)

			reqLogArgs := requestLogArgs{
				component:   "HTTP",
				header:      r.Header,
				fullMethod:  fullMethod,
				requestBody: string(reqBody),
			}
			if userName, err := types.GetUserKey(r.Context()); err == nil {
				reqLogArgs.user = userName
			}

			logRequestBeforeHandler(r.Context(), logger, reqLogArgs)

			start := time.Now()
			next.ServeHTTP(ww, r)
			took := time.Since(start)

			statusCodeInt := ww.StatusCode
			errType := httpRespErrorTypeFromStatusCode(statusCodeInt)

			reqLogArgs.statusCode = http.StatusText(statusCodeInt)
			reqLogArgs.took = took

			if errType == respClientError {
				reqLogArgs.clientError = ww.ErrorMessage
				if reqLogArgs.clientError == "" {
					reqLogArgs.clientError = "unknown error"
				}
			} else if errType == respServerError {
				reqLogArgs.serverError = ww.ErrorMessage
				if reqLogArgs.serverError == "" {
					reqLogArgs.serverError = "unknown error"
				}
			}

			logRequestAfterHandler(r.Context(), logger, reqLogArgs)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPMetricInterceptor() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			routePattern := getRoutePattern(ctx)

			metric.ServerRequestReceived.WithLabelValues("http", routePattern).Inc()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			took := time.Since(start)

			metric.HandledIncomingRequest(ctx, "http", routePattern, http.StatusText(ww.Status()), took)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPTraceInterceptor() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracing.StartSpan(r.Context(), getRoutePattern(r.Context()))
			defer span.End()

			r = r.WithContext(ctx)

			ww := httputil.NewWriter(w)

			next.ServeHTTP(ww, r)

			if ww.StatusCode >= 400 {
				span.SetAttributes(
					attribute.KeyValue{
						Key:   "error",
						Value: attribute.BoolValue(true),
					},
					attribute.KeyValue{
						Key:   "error_message",
						Value: attribute.StringValue(ww.ErrorMessage),
					},
					attribute.KeyValue{
						Key:   "status_code",
						Value: attribute.IntValue(ww.StatusCode),
					},
				)
			}
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPAuthInterceptor(providers *AuthProviders) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			noAuth := false

			if path, ok := swagger.GetSpecEx().FindPath(getRoutePattern(ctx)); ok {
				noAuth = !path.HasSecurity(r.Method)
			}

			if noAuth {
				next.ServeHTTP(w, r)
				return
			}

			uriApi, _, _, err := parseURI(r.RequestURI)
			if err != nil {
				msg := "failed to parse URI"
				logger.Error(msg, zap.Error(err))
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			_, checkAPIToken := tokenAuthServices[uriApi]
			username, err := providers.auth(ctx, authHeader, checkAPIToken)
			if err != nil {
				logger.Error("token auth failed", zap.Error(err))
				http.Error(w, "Unauthenticated", http.StatusUnauthorized)
				return
			}
			r = r.WithContext(context.WithValue(ctx, types.UserKey{}, username))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPRecoverInterceptor() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					fullMethod := strings.Join(
						[]string{req.Method, req.RequestURI, req.Proto}, " ",
					)
					handleRecover(fullMethod, r)
					http.Error(w, "recover: unexpected server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPRateLimitInterceptor(rateLimiters map[string]map[string]RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// methods without rate limiting = methods without auth
			noRateLimit := false

			if path, ok := swagger.GetSpecEx().FindPath(getRoutePattern(ctx)); ok {
				noRateLimit = !path.HasSecurity(r.Method)
			}

			if noRateLimit {
				next.ServeHTTP(w, r)
				return
			}

			uriApi, _, uriMethod, err := parseURI(r.RequestURI)
			if err != nil {
				logger.Error("failed to parse URI", zap.Error(err))
				http.NotFound(w, r)
				return
			}
			userToRateLimiter, ok := rateLimiters[uriApi]
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			limited, rlc, err := handleUserRateLimit(ctx, userToRateLimiter, uriMethod)
			if err != nil {
				logger.Error("failed to rate limit request", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeRateLimitHTTPHeaders(w, rlc)
			if limited {
				logger.Warn("request was rate limited")
				http.Error(w, "limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPNotFoundInterceptor() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.RawPath
			if path == "" {
				path = r.URL.Path
			}

			rCtx := chi.RouteContext(r.Context())
			tmpRouteCtx := chi.NewRouteContext()
			if !rCtx.Routes.Match(tmpRouteCtx, r.Method, path) {
				http.NotFound(w, r)
				return
			}

			r = r.WithContext(setRoutePattern(r.Context(), tmpRouteCtx.RoutePattern()))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func HTTPProcessHeadersInterceptor() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if header := r.Header.Get(types.UseSeqQLHeader); header != "" {
				r = r.WithContext(context.WithValue(r.Context(), types.UseSeqQL{}, header))
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
