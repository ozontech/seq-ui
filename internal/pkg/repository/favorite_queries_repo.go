package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ozontech/seq-ui/internal/app/types"
)

type favoriteQueriesRepository struct {
	*pool
}

func newFavoriteQueriesRepository(pool *pool) *favoriteQueriesRepository {
	return &favoriteQueriesRepository{pool}
}

func (r *favoriteQueriesRepository) GetAll(ctx context.Context, req types.GetFavoriteQueriesRequest) (types.FavoriteQueries, error) {
	query, args := "SELECT id, query, name, relative_from FROM favorite_queries WHERE profile_id = $1",
		[]any{req.ProfileID}

	metricLabels := []string{"favorite_queries", "SELECT"}
	rows, err := r.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get favorite queries: %w", err)
	}
	defer rows.Close()

	favoriteQueries := types.FavoriteQueries{}
	for rows.Next() {
		fQuery := types.FavoriteQuery{}
		if err = rows.Scan(&fQuery.ID, &fQuery.Query, &fQuery.Name, &fQuery.RelativeFrom); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		favoriteQueries = append(favoriteQueries, fQuery)
	}

	return favoriteQueries, nil
}

func (r *favoriteQueriesRepository) GetOrCreate(ctx context.Context, req types.GetOrCreateFavoriteQueryRequest) (int64, error) {
	var id int64 = -1

	query, args := "SELECT id FROM favorite_queries WHERE profile_id = $1 AND query = $2 AND name = $3 AND relative_from = $4 LIMIT 1",
		[]any{req.ProfileID, req.Query, req.Name, req.RelativeFrom}

	metricLabelsSelect := []string{"favorite_queries", "SELECT"}
	err := r.queryRow(ctx, metricLabelsSelect, query, args...).Scan(&id)

	// create favorite query if it doesn't exist
	if errors.Is(err, pgx.ErrNoRows) {
		query, args = "INSERT INTO favorite_queries (profile_id,query,name,relative_from) VALUES ($1,$2,$3,$4) RETURNING id",
			[]any{req.ProfileID, req.Query, req.Name, req.RelativeFrom}

		metricLabelsInsert := []string{"favorite_queries", "INSERT"}
		if err = r.queryRow(ctx, metricLabelsInsert, query, args...).Scan(&id); err != nil {
			incErrorMetric(err, metricLabelsInsert)
			return id, fmt.Errorf("failed to create favorite query: %w", err)
		}
	}
	if err != nil {
		incErrorMetric(err, metricLabelsSelect)
		return id, fmt.Errorf("failed to get favorite query: %w", err)
	}

	return id, nil
}

func (r *favoriteQueriesRepository) Delete(ctx context.Context, req types.DeleteFavoriteQueryRequest) error {
	query, args := "DELETE FROM favorite_queries WHERE id = $1 AND profile_id = $2",
		[]any{req.ID, req.ProfileID}

	metricLabels := []string{"favorite_queries", "DELETE"}
	if _, err := r.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to delete favorite query: %w", err)
	}

	return nil
}
