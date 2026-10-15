package seqapi_v1

import (
	"github.com/go-chi/chi/v5"

	grpc_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/seqapi/v1/http"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

type SeqAPI struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(cfg *config.SeqAPI, seqDBClients map[string]seqdb.Client, asyncSearches map[string]asyncsearches.Service) *SeqAPI {
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
