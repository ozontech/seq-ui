package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
)

type API struct {
	admin.UnimplementedAdminServiceServer

	service service.Service
}

func New(svc service.Service) *API {
	return &API{
		service: svc,
	}
}
