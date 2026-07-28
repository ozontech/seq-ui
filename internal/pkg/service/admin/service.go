package admin

import (
	"context"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
)

type Service interface {
	CreateRole(context.Context, types.CreateRoleRequest) (int32, error)
	AddUsersToRole(context.Context, types.AddUsersToRoleRequest) error
	DeleteUsersFromRole(context.Context, types.DeleteUsersFromRoleRequest) error
	GetRoles(context.Context) (types.GetRolesResponse, error)
	GetRole(context.Context, types.GetRoleRequest) (types.RoleInfo, error)
	UpdateRole(context.Context, types.UpdateRoleRequest) error
	DeleteRole(context.Context, types.DeleteRoleRequest) error
	GetAvailablePermissions() []types.PermissionGroup
}

type service struct {
	admin        repository.Admin
	userProfiles repository.UserProfiles
	cache        adminCache
	superUsers   map[string]struct{}
}

func New(admin repository.Admin, up repository.UserProfiles, c cache.Cache, cfg *config.Admin) Service {
	su := make(map[string]struct{}, len(cfg.SuperUsers))
	for _, u := range cfg.SuperUsers {
		su[u] = struct{}{}
	}

	return &service{
		admin:        admin,
		userProfiles: up,
		cache:        newAdminCache(c, cfg.CacheTTL),
		superUsers:   su,
	}
}
