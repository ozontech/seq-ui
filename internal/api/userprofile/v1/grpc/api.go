package grpc

import (
	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

type API struct {
	userprofile.UnimplementedUserProfileServiceServer

	service  service.Service
	profiles *profiles.Profiles
}

func New(svc service.Service, p *profiles.Profiles) *API {
	return &API{
		service:  svc,
		profiles: p,
	}
}
