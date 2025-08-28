package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	generated "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

type API struct {
	generated.UnimplementedErrorGroupsServiceServer

	service *errorgroups.Service
}

func New(svc *errorgroups.Service) *API {
	return &API{
		service: svc,
	}
}
