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

type permissionsCache struct {
	data map[string]permissionsCacheItem
	mu   sync.RWMutex
}

type permissionsCacheItem struct {
	permissions uint64
	expiresAt   time.Time
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

func newPermissionsCache() *permissionsCache {
	return &permissionsCache{
		data: make(map[string]permissionsCacheItem),
	}
}

func (c *permissionsCache) get(username string) (uint64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userInf, ok := c.data[username]
	if !ok || time.Now().After(userInf.expiresAt) {
		return 0, false
	}

	return userInf.permissions, true
}

func (c *permissionsCache) set(username string, permissions uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[username] = permissionsCacheItem{
		permissions: permissions,
		expiresAt:   time.Now().Add(permissionsCacheTTL),
	}
}

func (c *permissionsCache) reset(username string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, username)
}

func (c *permissionsCache) resetAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]permissionsCacheItem)
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
