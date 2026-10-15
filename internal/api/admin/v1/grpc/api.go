package grpc

import (
	"github.com/ozontech/seq-ui/internal/app/types"
	admin "github.com/ozontech/seq-ui/internal/pkg/service/admin"
	api "github.com/ozontech/seq-ui/pkg/admin/v1"
)

type API struct {
	api.UnimplementedAdminServiceServer

	services             map[string]admin.Service
	availablePermissions []*api.PermissionGroup
}

func New(services map[string]admin.Service) *API {
	// available permissions are the same for all envs
	var availablePermissions []*api.PermissionGroup
	for _, svc := range services {
		availablePermissions = availablePermissionsToProto(svc.GetAvailablePermissions())
		break
	}

	return &API{
		services:             services,
		availablePermissions: availablePermissions,
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
