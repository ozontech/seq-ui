package service

import (
	"context"

	"github.com/ozontech/seq-ui/internal/app/types"
)

// GetFavoriteQueries from underlying repository.
func (s *service) GetFavoriteQueries(ctx context.Context, req types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error) {
	return s.repo.FavoriteQueries.GetAll(ctx, req)
}

// GetOrCreateFavoriteQuery in underlying repository.
func (s *service) GetOrCreateFavoriteQuery(ctx context.Context, req types.GetOrCreateFavoriteQueryRequest) (int64, error) {
	if req.Query == "" {
		return -1, types.NewErrInvalidRequestField("empty query")
	}

	return s.repo.FavoriteQueries.GetOrCreate(ctx, req)
}

// DeleteFavoriteQuery in underlying repository.
func (s *service) DeleteFavoriteQuery(ctx context.Context, req types.DeleteFavoriteQueryRequest) error {
	if req.ID <= 0 {
		return types.NewErrInvalidRequestField("invalid id")
	}

	return s.repo.FavoriteQueries.Delete(ctx, req)
}
