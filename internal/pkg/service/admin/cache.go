package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
)

const (
	cacheKeyRoles     = "admin_roles"
	cacheKeyUserPerms = "admin_user_perms_"
)

type adminCache struct {
	cache cache.Cache
	ttl   time.Duration
}

func newAdminCache(c cache.Cache, ttl time.Duration) adminCache {
	return adminCache{cache: c, ttl: ttl}
}

func (c adminCache) getRoles(ctx context.Context) ([]types.Role, error) {
	data, err := c.cache.Get(ctx, cacheKeyRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles from cache: %w", err)
	}

	var roles []types.Role
	if err := json.Unmarshal([]byte(data), &roles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached roles: %w", err)
	}

	return roles, nil
}

func (c adminCache) setRoles(ctx context.Context, roles []types.Role) {
	data, err := json.Marshal(roles)
	if err != nil {
		logger.Error("failed to marshal roles for cache: %w", zap.Error(err))
	}

	if err := c.cache.SetWithTTL(ctx, cacheKeyRoles, string(data), c.ttl); err != nil {
		logger.Error("failed to set roles in cache: %w", zap.Error(err))
	}
}

func (c adminCache) resetRoles(ctx context.Context) {
	c.cache.Del(ctx, cacheKeyRoles)
}

func (c adminCache) getUserPermissions(ctx context.Context, username string) ([]string, error) {
	data, err := c.cache.Get(ctx, cacheKeyUserPerms+username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions from cache: %w", err)
	}

	var permissions []string
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached user permissions: %w", err)
	}

	return permissions, nil
}

func (c adminCache) setUserPermissions(ctx context.Context, username string, permissions []string) {
	data, err := json.Marshal(permissions)
	if err != nil {
		logger.Error("failed to marshal user permissions for cache: %w", zap.Error(err))
	}

	if err := c.cache.SetWithTTL(ctx, cacheKeyUserPerms+username, string(data), c.ttl); err != nil {
		logger.Error("failed to set user permissions in cache: %w", zap.Error(err))
	}
}

func (c adminCache) resetUsersPermissions(ctx context.Context, usernames ...string) {
	for _, username := range usernames {
		c.cache.Del(ctx, cacheKeyUserPerms+username)
	}
}
