package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	grpc_mw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ozontech/seq-ui/internal/api"
	"github.com/ozontech/seq-ui/internal/app/mw"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultCORSAllowedOrigins = "*"
	maxHTTPHeaderBytes        = 1 << 12 // 4 KiB
)

var defaultCORSAllowedMethods = []string{"HEAD", "GET", "POST", "PATCH", "DELETE"}

func (s *Server) init(ctx context.Context, registrar *api.Registrar) error {
	var err error

	s.authPrvds, err = mw.NewAuthProviders(ctx, s.config.JWTSecretKey, s.config.OIDC, s.config.Cache)
	if err != nil {
		return err
	}

	err = s.prepareRateLimiters()
	if err != nil {
		return err
	}

	s.prepareGRPCServer(registrar)

	s.prepareHTTPServer(ctx, registrar)

	err = s.prepareDebugServer(ctx)
	if err != nil {
		return nil
	}

	return nil
}

// setupCORS applies CORS policies set in config to the provided mux.
func (s *Server) setupCORS(mux *chi.Mux) {
	allowedOrigins := s.config.CORS.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, defaultCORSAllowedOrigins)
	}
	allowedMethods := s.config.CORS.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = append(allowedMethods, defaultCORSAllowedMethods...)
	}
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:     allowedOrigins,
		AllowedMethods:     allowedMethods,
		AllowedHeaders:     s.config.CORS.AllowedHeaders,
		ExposedHeaders:     s.config.CORS.ExposedHeaders,
		AllowCredentials:   s.config.CORS.AllowCredentials,
		MaxAge:             s.config.CORS.MaxAge,
		OptionsPassthrough: s.config.CORS.OptionsPassthrough,
	}))
}

// prepareRateLimiters prepares requests rate limiters based on server config.
func (s *Server) prepareRateLimiters() error {
	if len(s.config.RateLimiters) == 0 {
		return nil
	}

	s.rateLimiters = make(map[string]map[string]mw.RateLimiter)

	for apiName, rateLimiters := range s.config.RateLimiters {
		s.rateLimiters[apiName] = make(map[string]mw.RateLimiter)

		logger.Info("init default rate limiter", zap.String("api", apiName))
		defaultLimiter, err := mw.NewRateLimiter(apiName, rateLimiters.Default)
		if err != nil {
			return fmt.Errorf("init default rate limiter: %w", err)
		}

		s.rateLimiters[apiName][mw.RateLimiterDefaultUser] = defaultLimiter

		for user, rateLimiter := range rateLimiters.SpecialUsers {
			logger.Info(
				"init user rate limiter",
				zap.String("api", apiName),
				zap.String("user", user),
			)
			limiter, err := mw.NewRateLimiter(apiName, rateLimiter)
			if err != nil {
				return fmt.Errorf("init user %q rate limiter: %w", user, err)
			}

			s.rateLimiters[apiName][user] = limiter
		}
	}

	return nil
}

// prepareGRPCServer creates gRPC server and binds it to the Server instance.
func (s *Server) prepareGRPCServer(registrar *api.Registrar) {
	interceptors := []grpc.UnaryServerInterceptor{
		mw.GRPCRecoverInterceptor(),
		mw.GRPCProcessHeadersInterceptor(),
	}
	if s.authPrvds.JwtProvider != nil || s.authPrvds.OidcProvider != nil {
		interceptors = append(interceptors, mw.GRPCAuthInterceptor(&s.authPrvds))
	}
	interceptors = append(interceptors,
		mw.GRPCMetricInterceptor(),
		mw.GRPCTraceInterceptor(),
		mw.GRPCLogInterceptor(tracing.NewLogger(logger.Instance)),
	)
	if len(s.rateLimiters) > 0 {
		interceptors = append(interceptors, mw.GRPCRateLimitInterceptor(s.rateLimiters))
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_mw.ChainUnaryServer(interceptors...)),
		grpc.ConnectionTimeout(s.config.GRPCConnectionTimeout),
	}
	s.grpcServer = grpc.NewServer(opts...)

	registrar.RegisterGRPCHandlers(s.grpcServer)

	// support grpc clients like grpc_curl, evans, grpcui, etc
	reflection.Register(s.grpcServer)
}

// prepareHTTPServer prepares HTTP server with applied CORS policies and added interceptors.
func (s *Server) prepareHTTPServer(ctx context.Context, registrar *api.Registrar) {
	mux := chi.NewMux()
	if s.config.CORS != nil {
		s.setupCORS(mux)
	}
	interceptors := chi.Middlewares{
		mw.HTTPNotFoundInterceptor(),
		mw.HTTPRecoverInterceptor(),
		mw.HTTPProcessHeadersInterceptor(),
	}
	if s.authPrvds.JwtProvider != nil || s.authPrvds.OidcProvider != nil {
		interceptors = append(interceptors, mw.HTTPAuthInterceptor(&s.authPrvds))
	}
	interceptors = append(interceptors,
		mw.HTTPMetricInterceptor(),
		mw.HTTPTraceInterceptor(),
		mw.HTTPLogInterceptor(tracing.NewLogger(logger.Instance)),
	)
	if len(s.rateLimiters) > 0 {
		interceptors = append(interceptors, mw.HTTPRateLimitInterceptor(s.rateLimiters))
	}
	mux.Use(interceptors...)

	registrar.RegisterHTTPHandlers(mux)

	s.httpServer = s.makeHTTPServer(ctx, mux)
}

// prepareHTTPMux prepares debug HTTP server mux with applied CORS policies and added interceptors.
func (s *Server) prepareDebugServer(ctx context.Context) error {
	mux := chi.NewMux()
	if s.config.CORS != nil {
		s.setupCORS(mux)
	}
	interceptors := chi.Middlewares{
		mw.HTTPRecoverInterceptor(),
	}
	mux.Use(interceptors...)
	mux.Handle("/metrics", promhttp.Handler())
	serveHealth(mux)
	_, port, err := net.SplitHostPort(s.config.HTTPAddr)
	if err != nil {
		return err
	}
	serveSwaggerUI(mux, port)
	servePprof(mux)
	s.debugServer = s.makeHTTPServer(ctx, mux)
	return nil
}

// makeHTTPServer makes HTTP server from provided mux.
func (s *Server) makeHTTPServer(ctx context.Context, mux *chi.Mux) *http.Server {
	return &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: s.config.HTTPReadHeaderTimeout,
		ReadTimeout:       s.config.HTTPReadTimeout,
		WriteTimeout:      s.config.HTTPWriteTimeout,
		MaxHeaderBytes:    maxHTTPHeaderBytes, // 4 KiB
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}
