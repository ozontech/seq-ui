package admin

import (
	"context"
	"fmt"
	"slices"

	"github.com/ozontech/seq-ui/internal/app/types"
)

const (
	permissionCreateRoles = "roles:create"
	permissionReadRoles   = "roles:read"
	permissionUpdateRoles = "roles:update"
	permissionDeleteRoles = "roles:delete"
)

var availablePermissions = []types.PermissionGroup{
	{
		Group:       "roles",
		Permissions: []string{"create", "read", "update", "delete"},
	},
}

func (s *service) GetAvailablePermissions() []types.PermissionGroup {
	return availablePermissions
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

	availablePerms := parsePermissionGroupsToStrings(availablePermissions)

	for _, permission := range permissions {
		if !slices.Contains(availablePerms, permission) {
			return fmt.Errorf("unknown permission: %s", permission)
		}
	}

	return nil
}

func parsePermissionGroupsToStrings(groups []types.PermissionGroup) []string {
	lenPermStrs := 0
	for _, g := range groups {
		lenPermStrs += len(g.Permissions)
	}

	permStrs := make([]string, 0, lenPermStrs)
	for _, g := range groups {
		for _, p := range g.Permissions {
			permStrs = append(permStrs, g.Group+":"+p)
		}
	}

	return permStrs
}
