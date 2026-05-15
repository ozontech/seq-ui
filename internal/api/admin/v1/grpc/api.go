package grpc

import (
	"context"
	"strings"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
)

var grpcRoutePermissions = map[string]uint64{
	"Role": service.PermissionManageRoles,
}

type API struct {
	admin.UnimplementedAdminServiceServer

	service              service.Service
	availablePermissions []*admin.GetRolesResponse_Permission
	superUsers           map[string]struct{}
}

func New(svc service.Service, cfg *config.Admin) *API {
	su := make(map[string]struct{}, len(cfg.SuperUsers))
	for _, user := range cfg.SuperUsers {
		su[user] = struct{}{}
	}
	return &API{
		service:              svc,
		availablePermissions: availablePermissionsToProto(svc.GetAvailablePermissions()),
		superUsers:           su,
	}
}

func (a *API) authorize(ctx context.Context, method string) error {
	username, err := types.GetUserKey(ctx)
	if err != nil {
		return types.ErrUnauthenticated
	}

	if _, ok := a.superUsers[username]; ok {
		return nil
	}

	requiredPermission, ok := matchGRPCRoutePermissions(method)
	if !ok {
		return types.ErrPermissionDenied
	}

	permissions, err := a.service.GetUserPermissions(ctx, types.GetUserPermissionsRequest{Username: username})
	if err != nil {
		return types.ErrPermissionDenied
	}

	if permissions&requiredPermission == 0 {
		return types.ErrPermissionDenied
	}

	return nil
}

func matchGRPCRoutePermissions(method string) (uint64, bool) {
	for keyword, permission := range grpcRoutePermissions {
		if strings.Contains(method, keyword) {
			return permission, true
		}
	}
	return 0, false
}
