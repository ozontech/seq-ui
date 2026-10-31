package admin

import (
	"context"
	"time"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
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
}

type service struct {
	adminRepo        repository.Admin
	userProfilesRepo repository.UserProfiles
	cache            adminCache
	superUsers       map[string]struct{}
	env              string
}

func New(cfg *config.Admin, adminRepo repository.Admin, userProfilesRepo repository.UserProfiles) map[string]Service {
	serviceByEnv := make(map[string]Service, max(1, len(cfg.Envs)))
	if len(cfg.Envs) > 0 {
		for envName, envOptions := range cfg.Envs {
			serviceByEnv[envName] = newService(adminRepo, userProfilesRepo, envOptions, envName)
		}
	} else {
		serviceByEnv[""] = newService(adminRepo, userProfilesRepo, &cfg.Options, "")
	}

	return serviceByEnv
}

func newService(adminRepo repository.Admin, userProfilesRepo repository.UserProfiles, cfg *config.AdminOptions, env string) Service {
	su := make(map[string]struct{}, len(cfg.SuperUsers))
	for _, u := range cfg.SuperUsers {
		su[u] = struct{}{}
	}

	var (
		c   cache.Cache
		ttl time.Duration
	)
	if cfg.Cache != nil {
		c, _ = cache.ByID(cfg.Cache.ID)
		ttl = cfg.Cache.TTL
	}

	return &service{
		adminRepo:        adminRepo,
		userProfilesRepo: userProfilesRepo,
		cache:            newAdminCache(c, ttl, env),
		superUsers:       su,
		env:              env,
	}
}
