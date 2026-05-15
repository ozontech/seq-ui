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

type adminRepository struct {
	*pool
}

func newAdminRepository(pool *pool) *adminRepository {
	return &adminRepository{pool}
}

func (r *adminRepository) CreateRole(ctx context.Context, req types.CreateRoleRepoRequest) (int32, error) {
	query, args := "INSERT INTO roles (name, permissions) VALUES ($1, $2) RETURNING id", []any{req.Name, req.Permissions}
	metricLabels := []string{"admin", "INSERT"}

	var roleID int32
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&roleID); err != nil {
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	return roleID, nil
}

func (r *adminRepository) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	query, args := "UPDATE user_profiles SET role_id=$1 WHERE user_name=ANY($2)", []any{req.RoleID, req.Usernames}
	metricLabels := []string{"admin", "UPDATE"}

	if _, err := r.pool.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to add users to role: %w", err)
	}

	return nil
}

func (r *adminRepository) GetRoles(ctx context.Context) ([]types.RoleRepo, error) {
	query := "SELECT id, name, permissions FROM roles ORDER BY name"
	metricLabels := []string{"admin", "SELECT"}

	rows, err := r.pool.query(ctx, metricLabels, query)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	var roles []types.RoleRepo
	for rows.Next() {
		var role types.RoleRepo
		if err := rows.Scan(&role.ID, &role.Name, &role.Permissions); err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (r *adminRepository) GetRole(ctx context.Context, req types.GetRoleRequest) ([]types.Username, error) {
	metricLabels := []string{"admin", "SELECT"}
	query, args := "SELECT id FROM roles WHERE id=$1", []any{req.RoleID}

	var id int32
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, types.NewErrNotFound("role")
		}
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	query = "SELECT user_name FROM user_profiles WHERE role_id=$1 ORDER BY user_name"

	rows, err := r.pool.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}
	defer rows.Close()

	var usernames []types.Username
	for rows.Next() {
		var username types.Username
		if err := rows.Scan(&username); err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to scan role user: %w", err)
		}

		usernames = append(usernames, username)
	}

	return usernames, nil
}

func (r *adminRepository) UpdateRole(ctx context.Context, req types.UpdateRoleRepoRequest) error {
	metricLabels := []string{"admin", "UPDATE"}

	qb := sqlb.Update("roles").Where(sq.Eq{"id": req.RoleID})
	if req.Name != nil {
		qb = qb.Set("name", *req.Name)
	}
	if req.Permissions != nil {
		qb = qb.Set("permissions", *req.Permissions)
	}
	query, args := qb.MustSql()

	tag, err := r.pool.exec(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to update role: %w", err)
	}

	if tag.RowsAffected() == 0 {
		incErrorMetric(err, metricLabels)
		return types.NewErrNotFound("role")
	}

	return nil
}

func (r *adminRepository) DeleteRole(ctx context.Context, req types.DeleteRoleRequest) error {
	metricLabels := []string{"admin", "DELETE"}

	ctx, cancel := context.WithTimeout(ctx, r.pool.requestTimeout)
	defer cancel()

	tx, err := r.pool.pool.Begin(ctx)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to begin delete role transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query, args := "UPDATE user_profiles SET role_id=$1 WHERE role_id=$2", []any{0, req.RoleID}
	if req.ReplacementRoleID != nil {
		args[0] = *req.ReplacementRoleID
	}

	if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to update user_profiles when delete role: %w", err)
	}

	query, args = "DELETE FROM roles WHERE id=$1", []any{req.RoleID}
	tag, err := r.pool.execTx(ctx, tx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if tag.RowsAffected() == 0 {
		incErrorMetric(err, metricLabels)
		return types.NewErrNotFound("role")
	}

	if err := tx.Commit(ctx); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to commit delete role transaction: %w", err)
	}

	return nil
}

func (r *adminRepository) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) (uint64, error) {
	metricLabels := []string{"admin", "SELECT"}
	query, args := "SELECT r.permissions FROM user_profiles up JOIN roles r ON r.id=up.role_id WHERE up.user_name=$1", []any{req.Username}

	var permissions uint64
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&permissions); err != nil {
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}
