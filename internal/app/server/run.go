package server

import (
	"context"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/ozontech/seq-ui/logger"
)

// Run starts accepting new connections until the context is done.
func (s *Server) Run(ctx context.Context) error {
	errWg, ctx := errgroup.WithContext(ctx)

	// run gRPC server
	if s.grpcServer != nil {
		errWg.Go(func() error {
			lis, err := net.Listen("tcp", s.config.GRPC.Addr)
			if err != nil {
				return err
			}

			return s.grpcServer.Serve(lis)
		})
	}

	// run HTTP server
	if s.httpServer != nil {
		errWg.Go(func() error {
			l, err := net.Listen("tcp", s.config.HTTP.Addr)
			if err != nil {
				return err
			}

			return s.httpServer.Serve(l)
		})
	}

	// run debug server
	errWg.Go(func() error {
		l, err := net.Listen("tcp", s.config.Debug.Addr)
		if err != nil {
			return err
		}

		return s.debugServer.Serve(l)
	})

	errWg.Go(func() error {
		fields := []zap.Field{zap.String("debug", s.config.Debug.Addr)}
		if s.httpServer != nil {
			fields = append(fields,
				zap.String("http", s.config.HTTP.Addr),
				zap.String("swagger", s.config.Debug.Addr+swaggerUIPrefix),
			)
		}
		if s.grpcServer != nil {
			fields = append(fields, zap.String("grpc", s.config.GRPC.Addr))
		}

		logger.Info("app started", fields...)
		return nil
	})

	// graceful shutdown
	errWg.Go(func() error {
		<-ctx.Done()

		s.stop()

		return nil
	})

	return errWg.Wait()
}

// Stop the gRPC and HTTP servers.
func (s *Server) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var wg sync.WaitGroup
	if s.httpServer != nil {
		wg.Go(func() {
			if err := s.httpServer.Shutdown(ctx); err != nil {
				logger.Error("shutting down the http server", zap.Error(err))
			} else {
				logger.Warn("http server gracefully stopped")
			}
		})
	}

	wg.Go(func() {
		if err := s.debugServer.Shutdown(ctx); err != nil {
			logger.Error("shutting down the debug server", zap.Error(err))
		} else {
			logger.Warn("debug server gracefully stopped")
		}
	})

	if s.grpcServer != nil {
		done := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			logger.Warn("grpc server gracefully stopped")
		case <-ctx.Done():
			logger.Error("shutting down the grpc server", zap.Error(ctx.Err()))
			logger.Warn("grpc server force stop")
			s.grpcServer.Stop()
		}
	}
	wg.Wait()
}
