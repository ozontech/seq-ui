package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/dashboards"
	api "github.com/ozontech/seq-ui/pkg/dashboards/v1"
)

type API struct {
	api.UnimplementedDashboardsServiceServer

	service dashboards.Service
}

func New(svc dashboards.Service) *API {
	return &API{
		service: svc,
	}
}
