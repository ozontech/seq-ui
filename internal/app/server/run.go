package server

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/ozontech/seq-ui/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Run starts accepting new connections until the context is done.
func (s *Server) Run(ctx context.Context) error {
	errWg, ctx := errgroup.WithContext(ctx)

	// run gRPC server
	errWg.Go(func() error {
		lis, err := net.Listen("tcp", s.config.GRPCAddr)
		if err != nil {
			return err
		}

		return s.grpcServer.Serve(lis)
	})

	// run HTTP server
	errWg.Go(func() error {
		l, err := net.Listen("tcp", s.config.HTTPAddr)
		if err != nil {
			return err
		}

		return s.httpServer.Serve(l)
	})

	// run debug server
	errWg.Go(func() error {
		l, err := net.Listen("tcp", s.config.DebugAddr)
		if err != nil {
			return err
		}

		return s.debugServer.Serve(l)
	})

	errWg.Go(func() error {
		swaggerAddr := s.config.DebugAddr + swaggerUIPrefix
		logger.Info("app started",
			zap.String("http", s.config.HTTPAddr),
			zap.String("debug", s.config.DebugAddr),
			zap.String("swagger", swaggerAddr),
			zap.String("grpc", s.config.GRPCAddr),
		)
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
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	wg.Add(1)
	go func() {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Error("shutting down the http server", zap.Error(err))
		} else {
			logger.Warn("http server gracefully stopped")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		if err := s.debugServer.Shutdown(ctx); err != nil {
			logger.Error("shutting down the debug server", zap.Error(err))
		} else {
			logger.Warn("debug server gracefully stopped")
		}
		wg.Done()
	}()

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
	wg.Wait()
}
