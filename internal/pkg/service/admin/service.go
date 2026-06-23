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
	GetUserPermissions(context.Context, types.GetUserPermissionsRequest) ([]string, error)
	GetAvailablePermissions(context.Context) ([]types.Permission, error)
}

type service struct {
	repo               repository.Admin
	cache              cache.Cache
	availablePermCache *permissionsCache
	superUsers         map[string]struct{}
}

func New(repo repository.Admin, c cache.Cache, cfg *config.Admin) Service {
	su := make(map[string]struct{}, len(cfg.SuperUsers))
	for _, u := range cfg.SuperUsers {
		su[u] = struct{}{}
	}

	return &service{
		repo:               repo,
		cache:              c,
		availablePermCache: newPermissionsCache(),
		superUsers:         su,
	}
}
