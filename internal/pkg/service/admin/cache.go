package admin

import (
	"sync"

	"github.com/ozontech/seq-ui/internal/app/types"
)

type permissionsCache struct {
	availablePermissions []types.Permission
	mu                   sync.RWMutex
}

func newPermissionsCache() *permissionsCache {
	return &permissionsCache{}
}

func (c *permissionsCache) getAvailablePermissions() ([]types.Permission, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.availablePermissions, c.availablePermissions != nil
}

func (c *permissionsCache) setAvailablePermissions(permissions []types.Permission) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.availablePermissions = permissions
}
