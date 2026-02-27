package seqapi_v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	grpc_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/http"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

type SeqAPI struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(
	cfg config.SeqAPI,
	seqDB map[string]seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches *asyncsearches.Service,
	p *profiles.Profiles,
) *SeqAPI {
	return &SeqAPI{
		grpcAPI: grpc_api.New(cfg, seqDB, inmemWithRedisCache, redisCache, asyncSearches, p),
		httpAPI: http_api.New(cfg, seqDB, inmemWithRedisCache, redisCache, asyncSearches, p),
	}
}

func (s *SeqAPI) GRPCServer() *grpc_api.API {
	return s.grpcAPI
}

func (s *SeqAPI) HTTPRouter() chi.Router {
	return s.httpAPI.Router()
}
