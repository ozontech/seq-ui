package repository

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/ozontech/seq-ui/internal/app/types"
	sqlb "github.com/ozontech/seq-ui/internal/pkg/repository/sql_builder"
)

type asyncSearchesRepository struct {
	*pool
}

func newAsyncSearchesRepository(pool *pool) *asyncSearchesRepository {
	return &asyncSearchesRepository{pool}
}

func (r *asyncSearchesRepository) SaveAsyncSearch(
	ctx context.Context,
	req types.SaveAsyncSearchRequest,
) error {
	query, args := "INSERT INTO async_searches (search_id,owner_id,meta,expires_at) VALUES ($1,$2,$3,$4)",
		[]any{req.SearchID, req.OwnerID, req.Meta, req.ExpiresAt}

	metricLabels := []string{"async_searches", "INSERT"}
	_, err := r.exec(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to save async_search: %w", err)
	}

	return nil
}

func (r *asyncSearchesRepository) GetAsyncSearchById(
	ctx context.Context,
	searchID string,
) (types.AsyncSearchInfo, error) {
	as := types.AsyncSearchInfo{}

	query, args := `
		SELECT s.search_id, s.owner_id, p.user_name, s.meta, s.created_at, s.expires_at
		FROM async_searches AS s
		JOIN user_profiles AS p	ON p.id = s.owner_id
		WHERE s.search_id = $1
		`,
		[]any{searchID}

	metricLabels := []string{"async_searches", "SELECT"}
	err := r.queryRow(ctx, metricLabels, query, args...).Scan(
		&as.SearchID,
		&as.OwnerID,
		&as.OwnerName,
		&as.Meta,
		&as.CreatedAt,
		&as.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = types.NewErrNotFound("async_searches")
	} else if err != nil {
		err = fmt.Errorf("failed to get async search by id: %w", err)
	}

	if err != nil {
		incErrorMetric(err, metricLabels)
		return as, err
	}

	return as, nil
}

func (r *asyncSearchesRepository) GetAsyncSearchesList(
	ctx context.Context,
	req types.GetAsyncSearchesListRequest,
) ([]types.AsyncSearchInfo, error) {
	qb := sqlb.Select("s.search_id", "s.owner_id", "p.user_name", "s.created_at", "s.expires_at").
		From("async_searches AS s").
		Join("user_profiles AS p ON p.id = s.owner_id").
		Where(sq.GtOrEq{"s.expires_at": "now()"}).
		OrderBy("s.created_at DESC")

	if req.Owner != nil && *req.Owner != "" {
		qb = qb.Where(sq.Eq{
			"p.user_name": *req.Owner,
		})
	}
	if req.Limit > 0 {
		qb = qb.Limit(uint64(req.Limit))
	}
	if req.Offset > 0 {
		qb = qb.Offset(uint64(req.Offset))
	}

	query, args := qb.MustSql()

	metricLabels := []string{"async_searches", "SELECT"}
	rows, err := r.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get async searches: %w", err)
	}
	defer rows.Close()

	asyncSearches := make([]types.AsyncSearchInfo, 0)
	for rows.Next() {
		var as types.AsyncSearchInfo
		if err = rows.Scan(&as.SearchID, &as.OwnerID, &as.OwnerName, &as.CreatedAt, &as.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		asyncSearches = append(asyncSearches, as)
	}

	return asyncSearches, nil
}

func (r *asyncSearchesRepository) DeleteAsyncSearch(
	ctx context.Context,
	searchID string,
) error {
	query, args := "DELETE FROM async_searches WHERE search_id = $1",
		[]any{searchID}

	metricLabelsDelete := []string{"async_searches", "DELETE"}
	if _, err := r.exec(ctx, metricLabelsDelete, query, args...); err != nil {
		incErrorMetric(err, metricLabelsDelete)
		return fmt.Errorf("failed to delete async search: %w", err)
	}

	return nil
}

func (r *asyncSearchesRepository) DeleteExpiredAsyncSearches(ctx context.Context) error {
	query := "DELETE FROM async_searches WHERE expires_at < now()"

	metricLabelsDelete := []string{"async_searches", "DELETE"}
	if _, err := r.exec(ctx, metricLabelsDelete, query); err != nil {
		incErrorMetric(err, metricLabelsDelete)
		return fmt.Errorf("failed to delete expired async searches: %w", err)
	}

	return nil
}
