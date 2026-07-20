package admin

import (
	"context"
	"fmt"
	"slices"

	"github.com/ozontech/seq-ui/internal/app/types"
)

var availablePermissions = []string{
	permissionRolesCreate,
	permissionRolesRead,
	permissionRolesUpdate,
	permissionRolesDelete,
}

func (s *service) GetAvailablePermissions() []types.PermissionGroup {
	return []types.PermissionGroup{
		{
			Group:       "roles",
			Permissions: []string{"create", "read", "update", "delete"},
		},
	}
}

func (s *service) checkAccess(ctx context.Context, requiredPermission string) error {
	username, err := types.GetUserKey(ctx)
	if err != nil {
		return types.ErrUnauthenticated
	}

	if _, ok := s.superUsers[username]; ok {
		return nil
	}

	permissions, err := s.getUserPermissions(ctx, types.GetUserPermissionsRequest{Username: username})
	if err != nil {
		return fmt.Errorf("can't get user permissions: %w", err)
	}

	if !slices.Contains(permissions, requiredPermission) {
		return types.ErrPermissionDenied
	}

	return nil
}

func (s *service) getUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) ([]string, error) {
	if perms, err := s.cache.getUserPermissions(ctx, req.Username); err == nil {
		return perms, nil
	}

	perms, err := s.repo.GetUserPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.setUserPermissions(ctx, req.Username, perms)

	return perms, nil
}

func (s *service) validatePermissions(permissions []string) error {
	if len(permissions) == 0 {
		return types.NewErrInvalidRequestField("empty permissions")
	}

	for _, permission := range permissions {
		if !slices.Contains(availablePermissions, permission) {
			return fmt.Errorf("unknown permission: %s", permission)
		}
	}

	return nil
}
