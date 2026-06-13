package grpc

import (
	"github.com/ozontech/seq-ui/internal/app/types"
	adminservice "github.com/ozontech/seq-ui/internal/pkg/service/admin"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
)

type API struct {
	admin.UnimplementedAdminServiceServer

	service              adminservice.Service
	availablePermissions []*admin.GetRolesResponse_Permission
}

func New(svc adminservice.Service) *API {
	return &API{
		service:              svc,
		availablePermissions: availablePermissionsToProto(adminservice.GetAvailablePermissions()),
	}
}

func availablePermissionsToProto(source []types.Permission) []*admin.GetRolesResponse_Permission {
	availablePermissions := make([]*admin.GetRolesResponse_Permission, 0, len(source))
	for _, aPermission := range source {
		availablePermissions = append(availablePermissions, &admin.GetRolesResponse_Permission{
			Value:       aPermission.Value,
			Name:        aPermission.Name,
			Description: aPermission.Description,
		})
	}
	return availablePermissions
}
