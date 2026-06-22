package grpc

import (
	dashboardsservice "github.com/ozontech/seq-ui/internal/pkg/service/dashboards"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
)

type API struct {
	dashboards.UnimplementedDashboardsServiceServer

	service dashboardsservice.Service
}

func New(svc dashboardsservice.Service) *API {
	return &API{
		service: svc,
	}
}
