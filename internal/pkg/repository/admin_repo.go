package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ozontech/seq-ui/internal/app/types"
)

type adminRepository struct {
	*pool
}

func newAdminRepository(pool *pool) *adminRepository {
	return &adminRepository{pool}
}

func (r *adminRepository) CreateRole(ctx context.Context, req types.CreateRoleRepoRequest) (types.CreateRoleResponse, error) {
	query, args := "INSERT INTO roles (name, permissions) VALUES ($1, $2) RETURNING id", []any{req.Name, req.Permissions}

	var roleID int32
	metricLabels := []string{"admin", "CREATE"}
	resp := types.CreateRoleResponse{}

	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&roleID); err != nil {
		incErrorMetric(err, metricLabels)
		return resp, fmt.Errorf("failed to create role: %w", err)
	}

	resp.RoleID = roleID

	return resp, nil
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

func (r *adminRepository) GetRoles(ctx context.Context) (types.GetRolesRepoResponse, error) {
	query := "SELECT id, name, permissions FROM roles ORDER BY id"

	metricLabels := []string{"admin", "SELECT"}
	resp := types.GetRolesRepoResponse{}

	rows, err := r.pool.query(ctx, metricLabels, query)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return resp, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role types.RoleRepo
		if err := rows.Scan(&role.ID, &role.Name, &role.Permissions); err != nil {
			incErrorMetric(err, metricLabels)
			return resp, fmt.Errorf("failed to scan role: %w", err)
		}

		resp.Roles = append(resp.Roles, role)
	}

	return resp, nil
}

func (r *adminRepository) GetRole(ctx context.Context, req types.GetRoleRequest) (types.GetRoleResponse, error) {
	metricLabels := []string{"admin", "SELECT"}

	roleExists, err := r.roleExists(ctx, req.RoleID)
	if err != nil {
		return types.GetRoleResponse{}, err
	}
	if !roleExists {
		incErrorMetric(err, metricLabels)
		return types.GetRoleResponse{}, types.NewErrNotFound("role")
	}

	query, args := "SELECT user_name FROM user_profiles WHERE role_id=$1 ORDER BY user_name", []any{req.RoleID}
	resp := types.GetRoleResponse{}

	rows, err := r.pool.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return resp, fmt.Errorf("failed to get role users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			incErrorMetric(err, metricLabels)
			return resp, fmt.Errorf("failed to scan role user: %w", err)
		}

		resp.Usernames = append(resp.Usernames, username)
	}

	return resp, nil
}

func (r *adminRepository) UpdateRole(ctx context.Context, req types.UpdateRoleRepoRequest) error {
	metricLabels := []string{"admin", "UPDATE"}
	var (
		query string
		args  []any
	)

	switch {
	case req.Name != nil && req.Permissions != nil:
		query = "UPDATE roles SET name=$1, permissions=$2 WHERE id=$3"
		args = []any{*req.Name, *req.Permissions, req.RoleID}
	case req.Name != nil:
		query = "UPDATE roles SET name=$1 WHERE id=$2"
		args = []any{*req.Name, req.RoleID}
	default:
		query = "UPDATE roles SET permissions=$1 WHERE id=$2"
		args = []any{*req.Permissions, req.RoleID}
	}

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
	var (
		query string
		args  []any
	)

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

	if req.ReplacementRoleID != nil {
		replacementRoleExists, err := r.roleExistsTx(ctx, tx, *req.ReplacementRoleID)
		if err != nil {
			return err
		}
		if !replacementRoleExists {
			incErrorMetric(err, metricLabels)
			return types.NewErrNotFound("replacement role")
		}
	}

	switch {
	case req.ReplacementRoleID != nil:
		query = "UPDATE user_profiles SET role_id=$1 WHERE role_id=$2"
		args = []any{*req.ReplacementRoleID, req.RoleID}
	default:
		query = "UPDATE user_profiles SET role_id=NULL WHERE role_id=$1"
		args = []any{req.RoleID}
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

func (r *adminRepository) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) (types.GetUserPermissionsResponse, error) {
	metricLabels := []string{"admin", "SELECT"}
	query, args := "SELECT COALESCE(r.permissions, 0) FROM user_profiles up LEFT JOIN roles r ON r.id=up.role_id WHERE up.user_name=$1", []any{req.Username}

	var permissions uint64
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&permissions); err != nil {
		incErrorMetric(err, metricLabels)
		return types.GetUserPermissionsResponse{}, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return types.GetUserPermissionsResponse{Permissions: permissions}, nil
}

// Transactional check
func (r *adminRepository) roleExistsTx(ctx context.Context, tx pgx.Tx, roleID int32) (bool, error) {
	metricLabels := []string{"admin", "SELECT"}
	query, args := "SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)", []any{roleID}

	var exists bool
	if err := r.pool.queryRowTx(ctx, tx, metricLabels, query, args...).Scan(&exists); err != nil {
		incErrorMetric(err, metricLabels)
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return exists, nil
}

// Non-transactional check
func (r *adminRepository) roleExists(ctx context.Context, roleID int32) (bool, error) {
	metricLabels := []string{"admin", "SELECT"}
	query, args := "SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)", []any{roleID}

	var exists bool
	if err := r.pool.queryRow(ctx, metricLabels, query, args...).Scan(&exists); err != nil {
		incErrorMetric(err, metricLabels)
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return exists, nil
}
