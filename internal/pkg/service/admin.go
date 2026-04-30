package service

import (
	"context"

	"github.com/ozontech/seq-ui/internal/app/types"
)

func (s *service) CreateRole(ctx context.Context, req types.CreateRoleRequest) (types.CreateRoleResponse, error) {
	if req.Name == "" {
		return types.CreateRoleResponse{}, types.NewErrInvalidRequestField("empty role name")
	}

	if err := validatePermissions(req.Permissions); err != nil {
		return types.CreateRoleResponse{}, err
	}

	return s.repo.CreateRole(ctx, types.CreateRoleRepoRequest{
		Name:        req.Name,
		Permissions: permissionsToValue(req.Permissions),
	})
}

func (s *service) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if len(req.Usernames) == 0 {
		return types.NewErrInvalidRequestField("empty usernames")
	}

	for _, username := range req.Usernames {
		if username == "" {
			return types.NewErrInvalidRequestField("empty username")
		}
	}

	return s.repo.AddUsersToRole(ctx, req)
}

func (s *service) GetRoles(ctx context.Context) (types.GetRolesResponse, error) {
	resp, err := s.repo.GetRoles(ctx)
	if err != nil {
		return types.GetRolesResponse{}, err
	}

	roles := make([]types.Role, 0, len(resp.Roles))
	for _, role := range resp.Roles {
		roles = append(roles, types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: valueToPermissions(role.Permissions),
		})
	}

	return types.GetRolesResponse{
		Roles:                roles,
		AvailablePermissions: availablePermissions,
	}, nil
}

func (s *service) GetRole(ctx context.Context, req types.GetRoleRequest) (types.GetRoleResponse, error) {
	if req.RoleID <= 0 {
		return types.GetRoleResponse{}, types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	return s.repo.GetRole(ctx, req)
}

func (s *service) UpdateRole(ctx context.Context, req types.UpdateRoleRequest) error {
	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if req.Name != nil && *req.Name == "" {
		return types.NewErrInvalidRequestField("empty role name")
	}

	var permissions *uint64
	if len(req.Permissions) > 0 {
		if err := validatePermissions(req.Permissions); err != nil {
			return err
		}

		value := permissionsToValue(req.Permissions)
		permissions = &value
	}

	return s.repo.UpdateRole(ctx, types.UpdateRoleRepoRequest{
		RoleID:      req.RoleID,
		Name:        req.Name,
		Permissions: permissions,
	})
}

func (s *service) DeleteRole(ctx context.Context, req types.DeleteRoleRequest) error {
	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if req.ReplacementRoleID != nil {
		if *req.ReplacementRoleID <= 0 {
			return types.NewErrInvalidRequestField("value replacement_role_id must be greater than 0")
		}
		if *req.ReplacementRoleID == req.RoleID {
			return types.NewErrInvalidRequestField("replacement_role_id must be not equal to role_id")
		}
	}

	return s.repo.DeleteRole(ctx, req)
}
