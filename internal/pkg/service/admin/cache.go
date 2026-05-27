package admin

import (
	"sync"

	"github.com/ozontech/seq-ui/internal/app/types"
)

type adminCache struct {
	roles           []types.Role
	userPermissions map[string]uint64
	mu              sync.RWMutex
}

func newAdminCache() *adminCache {
	return &adminCache{
		userPermissions: make(map[string]uint64),
	}
}

func (c *adminCache) getRoles() ([]types.Role, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.roles, c.roles != nil
}

func (c *adminCache) getPermissions(username string) (perm uint64, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	perms, ok := c.userPermissions[username]
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

	c.userPermissions[username] = permissions
}

func (c *adminCache) resetRoles() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.roles = nil
}

func (c *adminCache) resetPermissions(usernames ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, un := range usernames {
		delete(c.userPermissions, un)
	}
}

func (c *adminCache) resetAllPermissions() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userPermissions = make(map[string]uint64)
}
