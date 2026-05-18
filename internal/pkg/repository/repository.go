package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/seq-ui/internal/app/types"
)

type (
	UserProfiles interface {
		GetOrCreate(context.Context, types.GetOrCreateUserProfileRequest) (types.UserProfile, error)
		Update(context.Context, types.UpdateUserProfileRequest) error
	}

	FavoriteQueries interface {
		GetAll(context.Context, types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error)
		GetOrCreate(context.Context, types.GetOrCreateFavoriteQueryRequest) (int64, error)
		Delete(context.Context, types.DeleteFavoriteQueryRequest) error
	}

	Dashboards interface {
		GetAll(context.Context, types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error)
		GetMy(context.Context, types.GetUserDashboardsRequest) (types.DashboardInfos, error)
		GetByUUID(context.Context, string) (types.Dashboard, error)
		Create(context.Context, types.CreateDashboardRequest) (string, error)
		Update(context.Context, types.UpdateDashboardRequest) error
		Delete(context.Context, types.DeleteDashboardRequest) error
		Search(context.Context, types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error)
	}

	AsyncSearches interface {
		SaveAsyncSearch(context.Context, types.SaveAsyncSearchRequest) error
		GetAsyncSearchById(context.Context, string) (types.AsyncSearchInfo, error)
		DeleteAsyncSearch(context.Context, string) error
		DeleteExpiredAsyncSearches(context.Context) error
		GetAsyncSearchesList(context.Context, types.GetAsyncSearchesListRequest) ([]types.AsyncSearchInfo, error)
	}

	Admin interface {
		CreateRole(context.Context, types.CreateRoleRepoRequest) (int32, error)
		AddUsersToRole(context.Context, types.AddUsersToRoleRequest) error
		GetRoles(context.Context) ([]types.RoleRepo, error)
		GetRole(context.Context, types.GetRoleRequest) ([]types.Username, error)
		UpdateRole(context.Context, types.UpdateRoleRepoRequest) error
		DeleteRole(context.Context, types.DeleteRoleRequest) error
		GetUserPermissions(context.Context, types.GetUserPermissionsRequest) (uint64, error)
	}
)

type Repository struct {
	UserProfiles
	FavoriteQueries
	Dashboards
	AsyncSearches
	Admin
}

func New(pool *pgxpool.Pool, requestTimeout time.Duration) *Repository {
	p := newPool(pool, requestTimeout)
	return &Repository{
		UserProfiles:    newUserProfilesRepository(p),
		FavoriteQueries: newFavoriteQueriesRepository(p),
		Dashboards:      newDashboardsRepository(p),
		AsyncSearches:   newAsyncSearchesRepository(p),
		Admin:           newAdminRepository(p),
	}
}
