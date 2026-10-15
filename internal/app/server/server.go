package server

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ozontech/seq-ui/internal/api"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/mw"
)

// Server contains application dependencies.
type Server struct {
	config      *config.Server
	debugServer *http.Server
	grpcServer  *grpc.Server
	httpServer  *http.Server

	envResolver  mw.EnvResolver
	authPrvds    mw.AuthProvidersByEnv
	rateLimiters mw.RateLimitersByEnv
}

// New returns a new Server.
func New(ctx context.Context, cfg *config.Config, registrar *api.Registrar) (*Server, error) {
	s := &Server{
		config:      &cfg.Server,
		envResolver: mw.NewEnvResolver(&cfg.Handlers),
	}

	if err := s.init(ctx, registrar); err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return s, nil
}
