package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/repository/txmanager"
)

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

func (r *adminRepository) CreateRole(ctx context.Context, req types.CreateRoleRequest) (int32, error) {
	var roleID int32
	err := r.txManager.Do(ctx, func(tx pgx.Tx) error {
		query, args := "INSERT INTO roles (name) VALUES ($1) RETURNING id", []any{req.Name}
		metricLabels := []string{"roles", "INSERT"}

		if err := r.pool.queryRowTx(ctx, tx, metricLabels, query, args...).Scan(&roleID); err != nil {
			incErrorMetric(err, metricLabels)
			return fmt.Errorf("failed to create role: %w", err)
		}

		if err := r.setRolePermissions(ctx, tx, roleID, req.Permissions); err != nil {
			return err
		}

		return nil
	})

	return roleID, err
}

func (r *adminRepository) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	metricLabels := []string{"users_roles", "INSERT"}
	query, args := `INSERT INTO users_roles (user_id, role_id)
		SELECT id, $1 FROM user_profiles WHERE user_name=ANY($2)
		ON CONFLICT DO NOTHING`, []any{req.RoleID, req.Usernames}

	if _, err := r.pool.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to add users to role: %w", err)
	}

	return nil
}

func (r *adminRepository) GetRoles(ctx context.Context) ([]types.Role, error) {
	metricLabels := []string{"roles", "SELECT"}
	query := `SELECT r.id, r.name, array_agg(p.value ORDER BY p.value) AS roles_permissions
		FROM roles r JOIN roles_permissions rp ON r.id=rp.role_id
		JOIN permissions p ON rp.permission_id=p.id GROUP BY r.id, r.name ORDER BY r.name`

	rows, err := r.pool.query(ctx, metricLabels, query)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	var roles []types.Role
	for rows.Next() {
		var role types.Role
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
	query, args := `SELECT up.user_name FROM user_profiles up JOIN users_roles ur ON up.id = ur.user_id
		WHERE ur.role_id=$1 ORDER BY up.user_name`, []any{req.RoleID}

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

func (r *adminRepository) UpdateRole(ctx context.Context, req types.UpdateRoleRequest) error {
	return r.txManager.Do(ctx, func(tx pgx.Tx) error {
		if req.Name != nil {
			metricLabels := []string{"roles", "UPDATE"}
			query, args := "UPDATE roles SET name=$1 WHERE id=$2", []any{*req.Name, req.RoleID}

			tag, err := r.pool.execTx(ctx, tx, metricLabels, query, args...)
			if err != nil {
				incErrorMetric(err, metricLabels)
				return fmt.Errorf("failed to update role name: %w", err)
			}

			if tag.RowsAffected() == 0 {
				return types.NewErrNotFound("role")
			}
		}

		if len(req.Permissions) > 0 {
			metricLabels := []string{"roles_permissions", "DELETE"}
			query, args := "DELETE FROM roles_permissions WHERE role_id=$1", []any{req.RoleID}

			if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
				incErrorMetric(err, metricLabels)
				return fmt.Errorf("failed to clear role permissions: %w", err)
			}

			if err := r.setRolePermissions(ctx, tx, req.RoleID, req.Permissions); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *adminRepository) DeleteRole(ctx context.Context, req types.DeleteRoleRequest) error {
	return r.txManager.Do(ctx, func(tx pgx.Tx) error {
		if req.ReplacementRoleID != nil {
			metricLabels := []string{"users_roles", "UPDATE"}
			query, args := "UPDATE users_roles SET role_id=$1 WHERE role_id=$2", []any{*req.ReplacementRoleID, req.RoleID}

			if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
				incErrorMetric(err, metricLabels)
				return fmt.Errorf("failed to reassign users: %w", err)
			}
		} else {
			metricLabels := []string{"users_roles", "DELETE"}
			query, args := "DELETE FROM users_roles WHERE role_id=$1", []any{req.RoleID}

			if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
				incErrorMetric(err, metricLabels)
				return fmt.Errorf("failed to remove users from role: %w", err)
			}
		}

		metricLabels := []string{"roles_permissions", "DELETE"}
		query, args := "DELETE FROM roles_permissions WHERE role_id=$1", []any{req.RoleID}

		if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
			incErrorMetric(err, metricLabels)
			return fmt.Errorf("failed to delete role permissions: %w", err)
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
	metricLabels := []string{"users_roles", "DELETE"}
	query, args := `DELETE FROM users_roles WHERE role_id=$1 AND user_id IN (
		SELECT id FROM user_profiles WHERE user_name=ANY($2))`, []any{req.RoleID, req.Usernames}

	if _, err := r.pool.exec(ctx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to delete users from role: %w", err)
	}

	return nil
}

func (r *adminRepository) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) ([]string, error) {
	metricLabels := []string{"permissions", "SELECT"}
	query, args := `SELECT DISTINCT p.value FROM user_profiles up
    	JOIN users_roles ur ON up.id=ur.user_id
     	JOIN roles_permissions rp ON ur.role_id=rp.role_id
      	JOIN permissions p ON rp.permission_id=p.id
       	WHERE up.user_name=$1 ORDER BY p.value`, []any{req.Username}

	rows, err := r.pool.query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (r *adminRepository) GetAvailablePermissions(ctx context.Context) ([]types.Permission, error) {
	metricLabels := []string{"permissions", "SELECT"}
	query := "SELECT id, value FROM permissions"

	rows, err := r.pool.query(ctx, metricLabels, query)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	defer rows.Close()

	var permissions []types.Permission
	for rows.Next() {
		var p types.Permission
		if err := rows.Scan(&p.ID, &p.Value); err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		permissions = append(permissions, p)
	}

	return permissions, nil
}

func (r *adminRepository) setRolePermissions(ctx context.Context, tx pgx.Tx, roleID int32, permissions []string) error {
	metricLabels := []string{"roles_permissions", "INSERT"}
	query, args := "INSERT INTO roles_permissions (role_id, permission_id) SELECT $1, id FROM permissions WHERE value=ANY($2)", []any{roleID, permissions}

	if _, err := r.pool.execTx(ctx, tx, metricLabels, query, args...); err != nil {
		incErrorMetric(err, metricLabels)
		return fmt.Errorf("failed to set role permissions: %w", err)
	}

	return nil
}
