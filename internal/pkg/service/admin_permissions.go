package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
)

const (
	PermissionManageRoles uint64 = 1 << iota
)

const permissionsCacheTTL = 30 * time.Second

type adminCache struct {
	roles       []types.Role
	permissions map[string]uint64
	mu          sync.RWMutex
}

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

func newAdminCache() *adminCache {
	return &adminCache{
		permissions: make(map[string]uint64),
	}
}

func (c *adminCache) getRoles() ([]types.Role, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.roles, c.roles != nil
}

func (c *adminCache) getPermissions(username string) (uint64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	perms, ok := c.permissions[username]
	if !ok {
		return 0, false
	}

	return perms, true
}

func (c *adminCache) setRoles(roles []types.Role) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.roles = roles
}

func (c *adminCache) setPermissions(username string, permissions uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.permissions[username] = permissions
}

func (c *adminCache) resetRoles() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.roles = nil
}

func (c *adminCache) resetPermissions(username string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.permissions, username)
}

func (c *adminCache) resetAllPermissions() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.permissions = make(map[string]uint64)
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
