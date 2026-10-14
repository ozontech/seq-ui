package seqapi_v1

import (
	"context"

	"github.com/go-chi/chi/v5"

	grpc_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/http"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

type SeqAPI struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(
	ctx context.Context,
	cfg *config.SeqAPI,
	seqDBClients map[string]seqdb.Client,
	asyncSearhCfg *config.AsyncSearch,
	asyncSearchesRepo repository.AsyncSearches,
) *SeqAPI {
	asyncSearches := newAsyncSearches(ctx, asyncSearhCfg, asyncSearchesRepo, seqDBClients)

	return &SeqAPI{
		grpcAPI: grpc_api.New(cfg, seqDBClients, asyncSearches),
		httpAPI: http_api.New(cfg, seqDBClients, asyncSearches),
	}
}

func (s *SeqAPI) GRPCServer() *grpc_api.API {
	return s.grpcAPI
}

func (s *SeqAPI) HTTPRouter() chi.Router {
	return s.httpAPI.Router()
}

func newAsyncSearches(
	ctx context.Context,
	cfg *config.AsyncSearch,
	repo repository.AsyncSearches,
	seqDBClients map[string]seqdb.Client,
) map[string]asyncsearches.Service {
	services := make(map[string]asyncsearches.Service, max(1, len(cfg.Envs)))
	if len(cfg.Envs) > 0 {
		for envName, envCfg := range cfg.Envs {
			services[envName] = asyncsearches.New(ctx, repo, seqDBClients[envCfg.SeqDBID], envCfg.Options, envName)
		}
	} else {
		var client seqdb.Client
		for _, c := range seqDBClients {
			client = c
		}
		services[""] = asyncsearches.New(ctx, repo, client, &cfg.Options, "")
	}

	return services
}
