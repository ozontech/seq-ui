package admin

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/types"
)

const (
	PermissionManageRoles uint64 = 1 << iota
)

var availablePermissions = []types.Permission{
	{
		Value:       PermissionManageRoles,
		Name:        "manage_roles",
		Description: "Manage roles",
	},
}

var availablePermissionsMap = map[uint64]struct{}{
	PermissionManageRoles: {},
}

//nolint:unparam
func (s *service) checkAccess(ctx context.Context, requiredPermission uint64) error {
	username, err := types.GetUserKey(ctx)
	if err != nil {
		return types.ErrUnauthenticated
	}

	if _, ok := s.superUsers[username]; ok {
		return nil
	}

	permissions, err := s.GetUserPermissions(ctx, types.GetUserPermissionsRequest{Username: username})
	if err != nil {
		return types.ErrPermissionDenied
	}

	if permissions&requiredPermission == 0 {
		return types.ErrPermissionDenied
	}

	return nil
}

func (s *service) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) (uint64, error) {
	if perms, ok := s.cache.getPermissions(req.Username); ok {
		return perms, nil
	}

	perms, err := s.repo.GetUserPermissions(ctx, req)
	if err != nil {
		return 0, err
	}

	s.cache.setPermissions(req.Username, perms)

	return perms, nil
}

func (s *service) GetAvailablePermissions() []types.Permission {
	return availablePermissions
}

func unmaskPermissions(value uint64) []uint64 {
	permissions := make([]uint64, 0)

	for _, permission := range availablePermissions {
		if value&permission.Value != 0 {
			permissions = append(permissions, permission.Value)
		}
	}

	return permissions
}

func validatePermissions(permissions []uint64) error {
	if len(permissions) == 0 {
		return types.NewErrInvalidRequestField("empty permissions")
	}

	for _, permission := range permissions {
		if _, ok := availablePermissionsMap[permission]; !ok {
			return fmt.Errorf("unknown permission: %d", permission)
		}
	}

	return nil
}

func maskPermissions(permissions []uint64) uint64 {
	var value uint64
	for _, permission := range permissions {
		value |= permission
	}
	return value
}
