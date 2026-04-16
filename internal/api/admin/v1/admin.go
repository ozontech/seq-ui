package admin_v1

import (
	"github.com/go-chi/chi/v5"
	grpc_api "github.com/ozontech/seq-ui/internal/api/admin/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/admin/v1/http"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type Admin struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(svc service.Service) *Admin {
	return &Admin{
		grpcAPI: grpc_api.New(svc),
		httpAPI: http_api.New(svc),
	}
}

func (a *Admin) GRPCServer() *grpc_api.API {
	return a.grpcAPI
}

func (a *Admin) HTTPRouter() chi.Router {
	return a.httpAPI.Router()
}
