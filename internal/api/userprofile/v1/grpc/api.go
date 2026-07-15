package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/userprofile"
	api "github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

type API struct {
	api.UnimplementedUserProfileServiceServer

	service userprofile.Service
}

func New(svc userprofile.Service) *API {
	return &API{
		service: svc,
	}
}
