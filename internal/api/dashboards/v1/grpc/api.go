package grpc

import (
	"github.com/ozontech/seq-ui/internal/api/profiles"
	dashboardsservice "github.com/ozontech/seq-ui/internal/pkg/service/dashboards"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
)

type API struct {
	dashboards.UnimplementedDashboardsServiceServer

	service  dashboardsservice.Service
	profiles *profiles.Profiles
}

func New(svc dashboardsservice.Service, p *profiles.Profiles) *API {
	return &API{
		service:  svc,
		profiles: p,
	}
}
