package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ozontech/seq-ui/internal/app/types"
	sqlb "github.com/ozontech/seq-ui/internal/pkg/repository/sql_builder"
)

type dashboardsRepository struct {
	*pool
}

func newDashboardsRepository(pool *pool) *dashboardsRepository {
	return &dashboardsRepository{pool}
}

func (r *dashboardsRepository) GetAll(ctx context.Context, req types.GetAllDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	query, args := `
		SELECT d.uuid, d.name, p.user_name
		FROM dashboards AS d
		JOIN user_profiles AS p	ON p.id = d.owner_id
		ORDER BY d.name ASC
		LIMIT $1
		OFFSET $2
		`,
		[]any{req.Limit, req.Offset}

	metricLabels := []string{"dashboards", "SELECT"}
	rows, err := r.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get all dashboards: %w", err)
	}
	defer rows.Close()

	dashboardInfos := types.DashboardInfosWithOwner{}
	for rows.Next() {
		var dashboardInfo types.DashboardInfoWithOwner
		if err = rows.Scan(&dashboardInfo.UUID, &dashboardInfo.Name, &dashboardInfo.OwnerName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dashboardInfos = append(dashboardInfos, dashboardInfo)
	}

	return dashboardInfos, nil
}

func (r *dashboardsRepository) GetMy(ctx context.Context, req types.GetUserDashboardsRequest) (types.DashboardInfos, error) {
	query, args := `
		SELECT uuid, name
		FROM dashboards
		WHERE owner_id = $1
		ORDER BY name ASC
		LIMIT $2
		OFFSET $3
		`,
		[]any{req.ProfileID, req.Limit, req.Offset}

	metricLabels := []string{"dashboards", "SELECT"}
	rows, err := r.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get user dashboards: %w", err)
	}
	defer rows.Close()

	dashboardInfos := types.DashboardInfos{}
	for rows.Next() {
		var dashboardInfo types.DashboardInfo
		if err = rows.Scan(&dashboardInfo.UUID, &dashboardInfo.Name); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dashboardInfos = append(dashboardInfos, dashboardInfo)
	}

	return dashboardInfos, nil
}

func (r *dashboardsRepository) GetByUUID(ctx context.Context, id string) (types.Dashboard, error) {
	dashboard := types.Dashboard{}

	query, args := `
		SELECT d.name, d.meta, p.user_name
		FROM dashboards AS d
		JOIN user_profiles AS p	ON p.id = d.owner_id
		WHERE d.uuid = $1
		LIMIT 1
		`,
		[]any{id}

	metricLabels := []string{"dashboards", "SELECT"}
	err := r.queryRow(ctx, metricLabels, query, args...).Scan(
		&dashboard.Name,
		&dashboard.Meta,
		&dashboard.OwnerName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = types.NewErrNotFound("dashboard")
	} else if err != nil {
		err = fmt.Errorf("failed to get dashboard: %w", err)
	}

	if err != nil {
		incErrorMetric(err, metricLabels)
		return dashboard, err
	}

	return dashboard, nil
}

func (r *dashboardsRepository) Create(ctx context.Context, req types.CreateDashboardRequest) (string, error) {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to create uuid v7: %w", err)
	}
	uuidStr := uuidv7.String()

	query, args := "INSERT INTO dashboards (uuid,owner_id,name,meta) VALUES ($1,$2,$3,$4)",
		[]any{uuidStr, req.ProfileID, req.Name, req.Meta}

	metricLabels := []string{"dashboards", "INSERT"}
	tag, err := r.exec(ctx, metricLabels, query, args...)
	if err != nil || tag.RowsAffected() == 0 {
		incErrorMetric(err, metricLabels)
		return "", fmt.Errorf("failed to create dashboard: %w", err)
	}

	return uuidStr, nil
}

func (r *dashboardsRepository) Update(ctx context.Context, req types.UpdateDashboardRequest) error {
	query, args := "SELECT owner_id FROM dashboards WHERE uuid = $1 LIMIT 1 FOR UPDATE",
		[]any{req.UUID}

	metricLabelsSelect := []string{"dashboards", "SELECT"}
	var ownerID int64
	err := r.queryRow(ctx, metricLabelsSelect, query, args...).Scan(&ownerID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err = types.NewErrNotFound("dashboard")
	case err != nil:
		err = fmt.Errorf("failed to get dashboard for update: %w", err)
	case ownerID != req.ProfileID:
		err = types.NewErrPermissionDenied("update dashboard")
	default:
		err = nil
	}

	if err != nil {
		incErrorMetric(err, metricLabelsSelect)
		return err
	}

	qb := sqlb.Update("dashboards").
		Where(sq.Eq{
			"uuid": req.UUID,
		})
	if req.Name != nil {
		qb = qb.Set("name", *req.Name)
	}
	if req.Meta != nil {
		qb = qb.Set("meta", *req.Meta)
	}

	query, args = qb.MustSql()

	metricLabelsUpdate := []string{"dashboards", "UPDATE"}
	if _, err = r.exec(ctx, metricLabelsUpdate, query, args...); err != nil {
		incErrorMetric(err, metricLabelsUpdate)
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	return nil
}

func (r *dashboardsRepository) Delete(ctx context.Context, req types.DeleteDashboardRequest) error {
	query, args := "SELECT owner_id FROM dashboards WHERE uuid = $1 LIMIT 1 FOR UPDATE",
		[]any{req.UUID}

	metricLabelsSelect := []string{"dashboards", "SELECT"}
	var ownerID int64
	err := r.queryRow(ctx, metricLabelsSelect, query, args...).Scan(&ownerID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err = nil
	case err != nil:
		err = fmt.Errorf("failed to get dashboard for delete: %w", err)
	case ownerID != req.ProfileID:
		err = types.NewErrPermissionDenied("delete dashboard")
	default:
		err = nil
	}

	if err != nil {
		incErrorMetric(err, metricLabelsSelect)
		return err
	}

	query, args = "DELETE FROM dashboards WHERE uuid = $1",
		[]any{req.UUID}

	metricLabelsDelete := []string{"dashboards", "DELETE"}
	if _, err = r.exec(ctx, metricLabelsDelete, query, args...); err != nil {
		incErrorMetric(err, metricLabelsDelete)
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	return nil
}

func (r *dashboardsRepository) Search(ctx context.Context, req types.SearchDashboardsRequest) (types.DashboardInfosWithOwner, error) {
	qb := sqlb.Select("d.uuid", "d.name", "p.user_name").
		From("dashboards AS d").
		Join("user_profiles AS p ON p.id = d.owner_id").
		Where(sq.Like{
			"LOWER(d.name)": fmt.Sprint("%", strings.ToLower(req.Query), "%"),
		}).
		OrderBy("d.name ASC").
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))
	if req.Filter != nil {
		if req.Filter.OwnerName != nil {
			qb = qb.Where(sq.Eq{
				"p.user_name": *req.Filter.OwnerName,
			})
		}
	}

	query, args := qb.MustSql()

	metricLabels := []string{"dashboards", "SELECT"}
	rows, err := r.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get dashboards: %w", err)
	}
	defer rows.Close()

	dashboardInfos := types.DashboardInfosWithOwner{}
	for rows.Next() {
		var dashboardInfo types.DashboardInfoWithOwner
		if err = rows.Scan(&dashboardInfo.UUID, &dashboardInfo.Name, &dashboardInfo.OwnerName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dashboardInfos = append(dashboardInfos, dashboardInfo)
	}

	return dashboardInfos, nil
}
