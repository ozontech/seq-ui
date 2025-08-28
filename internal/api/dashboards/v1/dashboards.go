package dashboards_v1

import (
	"github.com/go-chi/chi/v5"
	grpc_api "github.com/ozontech/seq-ui/internal/api/dashboards/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/dashboards/v1/http"
	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type Dashboards struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(svc service.Service, p *profiles.Profiles) *Dashboards {
	return &Dashboards{
		grpcAPI: grpc_api.New(svc, p),
		httpAPI: http_api.New(svc, p),
	}
}

func (d *Dashboards) GRPCServer() *grpc_api.API {
	return d.grpcAPI
}

func (d *Dashboards) HTTPRouter() chi.Router {
	return d.httpAPI.Router()
}
