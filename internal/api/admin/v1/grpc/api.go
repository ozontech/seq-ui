package grpc

import (
	admin "github.com/ozontech/seq-ui/internal/pkg/service/admin"
	api "github.com/ozontech/seq-ui/pkg/admin/v1"
)

type API struct {
	api.UnimplementedAdminServiceServer

	service admin.Service
}

func New(svc admin.Service) *API {
	return &API{
		service: svc,
	}
}
