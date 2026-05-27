package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ozontech/seq-ui/internal/app/types"
	sqlb "github.com/ozontech/seq-ui/internal/pkg/repository/sql_builder"
	"github.com/ozontech/seq-ui/internal/pkg/repository/txmanager"
)

const defaultRoleID = 0

type adminRepository struct {
	*pool
	txManager *txmanager.Manager
}

func newAdminRepository(pool *pool) *adminRepository {
	return &adminRepository{
		pool:      pool,
		txManager: txmanager.New(pool.pool),
	}
}

func (r *adminRepository) CreateRole(ctx context.Context, req types.CreateRoleRepoRequest) (int32, error) {
	query, args := "INSERT INTO roles (name, permissions) VALUES ($1, $2) RETURNING id", []any{req.Name, req.Permissions}
	metricLabels := []string{"roles", "INSERT"}

	var roleID int32
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&roleID); err != nil {
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to create role: %w", err)
	}

	return roleID, nil
}

func (r *adminRepository) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	query, args := "UPDATE user_profiles SET role_id=$1 WHERE user_name=ANY($2)", []any{req.RoleID, req.Usernames}
	metricLabels := []string{"user_profiles", "UPDATE"}

	if _, err := r.pool.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to add users to role: %w", err)
	}

	return nil
}

func (r *adminRepository) GetRoles(ctx context.Context) ([]types.RoleRepo, error) {
	query := "SELECT id, name, permissions FROM roles ORDER BY name"
	metricLabels := []string{"roles", "SELECT"}

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

func (r *adminRepository) GetRole(ctx context.Context, req types.GetRoleRequest) (types.RoleInfo, error) {
	metricLabels := []string{"user_profiles", "SELECT"}
	query, args := "SELECT user_name FROM user_profiles WHERE role_id=$1 ORDER BY user_name", []any{req.RoleID}

	rows, err := r.pool.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return types.RoleInfo{}, fmt.Errorf("failed to get role users: %w", err)
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			incErrorMetric(err, metricLabels)
			return types.RoleInfo{}, fmt.Errorf("failed to scan role user: %w", err)
		}

		usernames = append(usernames, username)
	}

	return types.RoleInfo{Usernames: usernames}, nil
}

func (r *adminRepository) UpdateRole(ctx context.Context, req types.UpdateRoleRepoRequest) error {
	metricLabels := []string{"roles", "UPDATE"}

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
		return types.NewErrNotFound("role")
	}

	return nil
}

func (r *adminRepository) DeleteRole(ctx context.Context, req types.DeleteRoleRequest) error {
	metricLabels := []string{"user_profiles", "UPDATE"}

	return r.txManager.Do(ctx, func(tx pgx.Tx) error {
		query, args := "UPDATE user_profiles SET role_id=$1 WHERE role_id=$2", []any{defaultRoleID, req.RoleID}
		if req.ReplacementRoleID != nil {
			args[0] = *req.ReplacementRoleID
		}

		if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
			incErrorMetric(err, metricLabels)
			return fmt.Errorf("failed to update user_profiles when delete role: %w", err)
		}

		metricLabels = []string{"roles", "DELETE"}
		query, args = "DELETE FROM roles WHERE id=$1", []any{req.RoleID}
		tag, err := r.pool.execTx(ctx, tx, metricLabels, query, args...)
		if err != nil {
			incErrorMetric(err, metricLabels)
			return fmt.Errorf("failed to delete role: %w", err)
		}

		if tag.RowsAffected() == 0 {
			return types.NewErrNotFound("role")
		}

		return nil
	})
}

func (r *adminRepository) DeleteUsersFromRole(ctx context.Context, req types.DeleteUsersFromRoleRequest) error {
	metricLabels := []string{"user_profiles", "UPDATE"}
	query, args := "UPDATE user_profiles SET role_id=$1 WHERE role_id=$2 AND user_name=ANY($3)", []any{defaultRoleID, req.RoleID, req.Usernames}

	if _, err := r.pool.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to delete users from role: %w", err)
	}

	return nil
}

func (r *adminRepository) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) (uint64, error) {
	metricLabels := []string{"user_profiles", "SELECT"}
	query, args := "SELECT r.permissions FROM user_profiles up JOIN roles r ON r.id=up.role_id WHERE up.user_name=$1", []any{req.Username}

	var permissions uint64
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&permissions); err != nil {
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}
