package service

import (
	"fmt"

	"github.com/ozontech/seq-ui/internal/app/types"
)

const (
	PermissionAdminAccess uint64 = 1 << iota
	PermissionManageRoles
)

var availablePermissions = []types.Permission{
	{
		Value:       PermissionAdminAccess,
		Name:        "admin_access",
		Description: ptr("Access to admin page"),
	},
	{
		Value:       PermissionManageRoles,
		Name:        "manage_roles_access",
		Description: ptr("Manage roles"),
	},
}

var availablePermissionsMap = map[uint64]struct{}{
	PermissionAdminAccess: {},
	PermissionManageRoles: {},
}

func valueToPermissions(value uint64) []uint64 {
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

func permissionsToValue(permissions []uint64) uint64 {
	var value uint64
	for _, permission := range permissions {
		value |= permission
	}
	return value
}

func ptr(desc string) *string {
	return &desc
}
