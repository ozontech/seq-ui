package grpc

import (
	userprofilesservice "github.com/ozontech/seq-ui/internal/pkg/service/userprofile"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

type API struct {
	userprofile.UnimplementedUserProfileServiceServer

	service userprofilesservice.Service
}

func New(svc userprofilesservice.Service) *API {
	return &API{
		service: svc,
	}
}
