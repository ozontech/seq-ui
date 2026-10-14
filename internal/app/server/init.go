package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	grpc_mw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ozontech/seq-ui/internal/api"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/mw"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/tracing"
)

func (s *Server) init(ctx context.Context, registrar *api.Registrar) error {
	if err := s.prepareAuth(ctx); err != nil {
		return err
	}
	if err := s.prepareRateLimiters(); err != nil {
		return err
	}

	s.prepareGRPCServer(registrar)
	s.prepareHTTPServer(ctx, registrar)

	if err := s.prepareDebugServer(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) prepareAuth(ctx context.Context) error {
	if s.config.Auth == nil {
		return nil
	}

	s.authPrvds = make(mw.AuthProvidersByEnv, len(s.config.Auth.Envs)+1)

	provider, err := mw.NewAuthProvider(ctx, s.config.Auth.Options)
	if err != nil {
		return fmt.Errorf("init auth: %w", err)
	}
	s.authPrvds[""] = provider

	for envName, envConfig := range s.config.Auth.Envs {
		provider, err := mw.NewAuthProvider(ctx, envConfig)
		if err != nil {
			return fmt.Errorf("init auth for env %q: %w", envName, err)
		}
		s.authPrvds[envName] = provider
	}
	return nil
}

// setupCORS applies CORS policies set in config to the provided mux.
func (s *Server) setupCORS(mux *chi.Mux, cfg *config.CORS) {
	if cfg == nil {
		return
	}

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:     cfg.AllowedOrigins,
		AllowedMethods:     cfg.AllowedMethods,
		AllowedHeaders:     cfg.AllowedHeaders,
		ExposedHeaders:     cfg.ExposedHeaders,
		AllowCredentials:   cfg.AllowCredentials,
		MaxAge:             cfg.MaxAge,
		OptionsPassthrough: cfg.OptionsPassthrough,
	}))
}

// prepareRateLimiters prepares requests rate limiters based on server config.
func (s *Server) prepareRateLimiters() error {
	if s.config.RateLimiting == nil {
		return nil
	}

	s.rateLimiters = make(mw.RateLimitersByEnv, len(s.config.RateLimiting.Envs)+1)

	limiters, err := mw.NewRateLimiters(s.config.RateLimiting.Options)
	if err != nil {
		return fmt.Errorf("init rate limiter: %w", err)
	}
	s.rateLimiters[""] = limiters

	for envName, envConfig := range s.config.RateLimiting.Envs {
		limiters, err := mw.NewRateLimiters(envConfig)
		if err != nil {
			return fmt.Errorf("init rate limiter for env %q: %w", envName, err)
		}
		s.rateLimiters[envName] = limiters
	}

	return nil
}

// prepareGRPCServer creates gRPC server and binds it to the Server instance.
func (s *Server) prepareGRPCServer(registrar *api.Registrar) {
	if s.config.GRPC == nil {
		return
	}

	interceptors := []grpc.UnaryServerInterceptor{
		mw.GRPCRecoverInterceptor(),
		mw.GRPCProcessHeadersInterceptor(),
		mw.GRPCEnvInterceptor(s.envResolver),
	}
	if len(s.authPrvds) > 0 {
		interceptors = append(interceptors, mw.GRPCAuthInterceptor(s.authPrvds))
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
		grpc.ConnectionTimeout(s.config.GRPC.ConnectionTimeout),
	}
	s.grpcServer = grpc.NewServer(opts...)

	registrar.RegisterGRPCHandlers(s.grpcServer)

	// support grpc clients like grpc_curl, evans, grpcui, etc
	reflection.Register(s.grpcServer)
}

// prepareHTTPServer prepares HTTP server with applied CORS policies and added interceptors.
func (s *Server) prepareHTTPServer(ctx context.Context, registrar *api.Registrar) {
	if s.config.HTTP == nil {
		return
	}

	mux := chi.NewMux()
	s.setupCORS(mux, s.config.HTTP.CORS)

	interceptors := chi.Middlewares{
		mw.HTTPNotFoundInterceptor(),
		mw.HTTPRecoverInterceptor(),
		mw.HTTPProcessHeadersInterceptor(),
		mw.HTTPEnvInterceptor(s.envResolver),
	}
	if len(s.authPrvds) > 0 {
		interceptors = append(interceptors, mw.HTTPAuthInterceptor(s.authPrvds))
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

	s.httpServer = s.makeHTTPServer(ctx, mux, s.config.HTTP.ReadHeaderTimeout, s.config.HTTP.ReadTimeout, s.config.HTTP.WriteTimeout)
}

// prepareHTTPMux prepares debug HTTP server mux with applied CORS policies and added interceptors.
func (s *Server) prepareDebugServer(ctx context.Context) error {
	mux := chi.NewMux()
	s.setupCORS(mux, s.config.Debug.CORS)
	interceptors := chi.Middlewares{
		mw.HTTPRecoverInterceptor(),
	}
	mux.Use(interceptors...)
	mux.Handle("/metrics", promhttp.Handler())
	serveHealth(mux)
	// swagger describes http api, so it is served only if http server is configured
	if s.config.HTTP != nil {
		_, port, err := net.SplitHostPort(s.config.HTTP.Addr)
		if err != nil {
			return err
		}
		serveSwaggerUI(mux, port)
	}
	servePprof(mux)
	s.debugServer = s.makeHTTPServer(ctx, mux, s.config.Debug.ReadHeaderTimeout, s.config.Debug.ReadTimeout, s.config.Debug.WriteTimeout)
	return nil
}

// makeHTTPServer makes HTTP server from provided mux.
func (s *Server) makeHTTPServer(ctx context.Context, mux *chi.Mux, readHeaderTimeout, readTimeout, writeTimeout time.Duration) *http.Server {
	return &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		MaxHeaderBytes:    config.MaxHTTPHeaderBytes, // 4 KiB
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}
