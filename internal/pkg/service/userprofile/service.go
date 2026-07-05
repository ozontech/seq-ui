package userprofile

import (
	"context"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/internal/pkg/service/profiles"
)

type Service interface {
	GetOrCreateUserProfile(context.Context, types.GetOrCreateUserProfileRequest) (types.UserProfile, error)
	UpdateUserProfile(context.Context, types.UpdateUserProfileRequest) error
	GetFavoriteQueries(context.Context, types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error)
	GetOrCreateFavoriteQuery(context.Context, types.GetOrCreateFavoriteQueryRequest) (int64, error)
	DeleteFavoriteQuery(context.Context, types.DeleteFavoriteQueryRequest) error
}

type service struct {
	UserProfiles    repository.UserProfiles
	FavoriteQueries repository.FavoriteQueries
}

func New(up repository.UserProfiles, fq repository.FavoriteQueries) Service {
	return &service{
		UserProfiles:    up,
		FavoriteQueries: fq,
	}
}

func (s *service) GetOrCreateUserProfile(ctx context.Context, req types.GetOrCreateUserProfileRequest) (types.UserProfile, error) {
	up, err := s.UserProfiles.GetOrCreate(ctx, req)
	if err != nil {
		return up, err
	}

	return up, nil
}

func (s *service) UpdateUserProfile(ctx context.Context, req types.UpdateUserProfileRequest) error {
	if req.IsEmpty() {
		return types.ErrEmptyUpdateRequest
	}
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return types.NewErrInvalidRequestField("invalid timezone format")
		}
	}

	return s.UserProfiles.Update(ctx, req)
}

func (s *service) GetFavoriteQueries(ctx context.Context, req types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error) {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.ProfileID = profileID

	return s.FavoriteQueries.GetAll(ctx, req)
}

func (s *service) GetOrCreateFavoriteQuery(ctx context.Context, req types.GetOrCreateFavoriteQueryRequest) (int64, error) {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	req.ProfileID = profileID

	if req.Query == "" {
		return -1, types.NewErrInvalidRequestField("empty query")
	}

	return s.FavoriteQueries.GetOrCreate(ctx, req)
}

func (s *service) DeleteFavoriteQuery(ctx context.Context, req types.DeleteFavoriteQueryRequest) error {
	profileID, err := profiles.GetIDFromContext(ctx)
	if err != nil {
		return err
	}
	req.ProfileID = profileID

	if req.ID <= 0 {
		return types.NewErrInvalidRequestField("invalid id")
	}

	return s.FavoriteQueries.Delete(ctx, req)
}
