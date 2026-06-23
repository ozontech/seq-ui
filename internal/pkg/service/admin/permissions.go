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

func (s *service) checkAccess(ctx context.Context, requiredPermission string) error {
	username, err := types.GetUserKey(ctx)
	if err != nil {
		return types.ErrUnauthenticated
	}

	if _, ok := s.superUsers[username]; ok {
		return nil
	}

	permissions, err := s.GetUserPermissions(ctx, types.GetUserPermissionsRequest{Username: username})
	if err != nil {
		return fmt.Errorf("can't get user permissions: %w", err)
	}

	if !slices.Contains(permissions, requiredPermission) {
		return types.ErrPermissionDenied
	}

	return nil
}

func (s *service) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) ([]string, error) {
	if perms, ok := s.cache.getPermissions(req.Username); ok {
		return perms, nil
	}

	perms, err := s.repo.GetUserPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.setPermissions(req.Username, perms)

	return perms, nil
}

func (s *service) GetAvailablePermissions(ctx context.Context) ([]types.Permission, error) {
	if perms, ok := s.cache.getAvailablePermissions(); ok {
		return perms, nil
	}

	perms, err := s.repo.GetAvailablePermissions(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.setAvailablePermissions(perms)

	return perms, nil
}

func (s *service) validatePermissions(ctx context.Context, permissions []string) error {
	if len(permissions) == 0 {
		return types.NewErrInvalidRequestField("empty permissions")
	}

	available, err := s.GetAvailablePermissions(ctx)
	if err != nil {
		return fmt.Errorf("can't get available permissions: %w", err)
	}

	for _, permission := range permissions {
		if !slices.ContainsFunc(available, func(aPermission types.Permission) bool {
			return aPermission.Value == permission
		}) {
			return fmt.Errorf("unknown permission: %s", permission)
		}
	}

	return nil
}
