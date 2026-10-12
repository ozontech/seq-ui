package server

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ozontech/seq-ui/internal/api"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/mw"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
)

// Server contains application dependencies.
type Server struct {
	config      *config.Server
	debugServer *http.Server
	grpcServer  *grpc.Server
	httpServer  *http.Server
	caches      map[string]cache.Cache

	envResolver  mw.EnvResolver
	authPrvds    mw.AuthProvidersByEnv
	rateLimiters mw.RateLimitersByEnv
}

// New returns a new Server.
func New(ctx context.Context, cfg *config.Config, registrar *api.Registrar, caches map[string]cache.Cache) (*Server, error) {
	s := &Server{
		config:      &cfg.Server,
		caches:      caches,
		envResolver: mw.NewEnvResolver(&cfg.Handlers),
	}

	if err := s.init(ctx, registrar); err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return s, nil
}
