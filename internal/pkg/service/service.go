package service

import (
	"context"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
)

type Service interface {
	GetOrCreateUserProfile(context.Context, types.GetOrCreateUserProfileRequest) (types.UserProfile, error)
	UpdateUserProfile(context.Context, types.UpdateUserProfileRequest) error

	GetFavoriteQueries(context.Context, types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error)
	GetOrCreateFavoriteQuery(context.Context, types.GetOrCreateFavoriteQueryRequest) (int64, error)
	DeleteFavoriteQuery(context.Context, types.DeleteFavoriteQueryRequest) error

	GetAllDashboards(context.Context, types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error)
	GetMyDashboards(context.Context, types.GetUserDashboardsRequest) (types.DashboardInfos, error)
	GetDashboardByUUID(context.Context, string) (types.Dashboard, error)
	CreateDashboard(context.Context, types.CreateDashboardRequest) (string, error)
	UpdateDashboard(context.Context, types.UpdateDashboardRequest) error
	DeleteDashboard(context.Context, types.DeleteDashboardRequest) error
	SearchDashboards(context.Context, types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error)

	CreateRole(context.Context, types.CreateRoleRequest) (int32, error)
	AddUsersToRole(context.Context, types.AddUsersToRoleRequest) error
	GetRoles(context.Context) (types.GetRolesResponse, error)
	GetRole(context.Context, types.GetRoleRequest) ([]types.Username, error)
	UpdateRole(context.Context, types.UpdateRoleRequest) error
	DeleteRole(context.Context, types.DeleteRoleRequest) error
	GetUserPermissions(context.Context, types.GetUserPermissionsRequest) (uint64, error)
	GetAvailablePermissions() []types.Permission
}

type service struct {
	repo      *repository.Repository
	permCache *permissionsCache
}

func New(repo *repository.Repository) Service {
	return &service{
		repo:      repo,
		permCache: newPermissionsCache(),
	}
}
