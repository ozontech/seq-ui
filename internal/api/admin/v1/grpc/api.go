package grpc

import (
	"github.com/ozontech/seq-ui/internal/app/types"
	admin "github.com/ozontech/seq-ui/internal/pkg/service/admin"
	api "github.com/ozontech/seq-ui/pkg/admin/v1"
)

type API struct {
	api.UnimplementedAdminServiceServer

	service              admin.Service
	availablePermissions []*api.PermissionGroup
}

func New(svc admin.Service) *API {
	return &API{
		service:              svc,
		availablePermissions: availablePermissionsToProto(svc.GetAvailablePermissions()),
	}
}

func availablePermissionsToProto(source []types.PermissionGroup) []*api.PermissionGroup {
	availablePermissions := make([]*api.PermissionGroup, 0, len(source))
	for _, s := range source {
		availablePermissions = append(availablePermissions, &api.PermissionGroup{
			Group:       s.Group,
			Permissions: s.Permissions,
		})
	}
	return availablePermissions
}
