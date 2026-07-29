package server

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ozontech/seq-ui/internal/api"
	config "github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/mw"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
)

// Server contains application dependencies.
type Server struct {
	config      *config.Server
	oidcCache   cache.Cache
	debugServer *http.Server
	grpcServer  *grpc.Server
	httpServer  *http.Server

	authPrvds    mw.AuthProviders
	rateLimiters map[string]map[string]mw.RateLimiter // rate limiter by api and user
}

// New returns a new Server.
func New(ctx context.Context, cfg *config.Server, registrar *api.Registrar, oidcCache cache.Cache) (*Server, error) {
	s := &Server{config: cfg, oidcCache: oidcCache}

	if err := s.init(ctx, registrar); err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return s, nil
}
