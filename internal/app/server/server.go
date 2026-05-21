package server

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ozontech/seq-ui/internal/api"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/mw"
)

// Server contains application dependencies.
type Server struct {
	config      *config.Server
	debugServer *http.Server
	grpcServer  *grpc.Server
	httpServer  *http.Server

	authPrvds    mw.AuthProviders
	rateLimiters map[string]map[string]mw.RateLimiter // rate limiter by api and user
}

// New returns a new Server.
func New(ctx context.Context, cfg *config.Server, registrar *api.Registrar) (*Server, error) {
	s := &Server{config: cfg}

	if err := s.init(ctx, registrar); err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return s, nil
}
