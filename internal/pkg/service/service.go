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
}

type service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) Service {
	return &service{
		repo: repo,
	}
}
