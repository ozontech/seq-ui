package errorgroups_v1

import (
	"github.com/go-chi/chi/v5"

	grpc_api "github.com/ozontech/seq-ui/internal/api/errorgroups/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/errorgroups/v1/http"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
)

type ErrorGroups struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(service *errorgroups.Service) *ErrorGroups {
	return &ErrorGroups{
		grpcAPI: grpc_api.New(service),
		httpAPI: http_api.New(service),
	}
}

func (s *ErrorGroups) GRPCServer() *grpc_api.API {
	return s.grpcAPI
}

func (s *ErrorGroups) HTTPRouter() chi.Router {
	return s.httpAPI.Router()
}
