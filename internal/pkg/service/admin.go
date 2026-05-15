package service

import (
	"context"
	"slices"

	"github.com/ozontech/seq-ui/internal/app/types"
)

func (s *service) CreateRole(ctx context.Context, req types.CreateRoleRequest) (int32, error) {
	if req.Name == "" {
		return 0, types.NewErrInvalidRequestField("empty role name")
	}

	if err := validatePermissions(req.Permissions); err != nil {
		return 0, err
	}

	roleID, err := s.repo.CreateRole(ctx, types.CreateRoleRepoRequest{
		Name:        req.Name,
		Permissions: maskPermissions(req.Permissions),
	})
	if err != nil {
		return 0, err
	}

	s.adminCache.resetRoles()

	return roleID, nil
}

func (s *service) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if len(req.Usernames) == 0 {
		return types.NewErrInvalidRequestField("empty usernames")
	}

	if slices.Contains(req.Usernames, "") {
		return types.NewErrInvalidRequestField("empty username")
	}

	if err := s.repo.AddUsersToRole(ctx, req); err != nil {
		return err
	}

	for _, username := range req.Usernames {
		s.adminCache.resetPermissions(username)
	}

	return nil
}

func (s *service) GetRoles(ctx context.Context) (types.GetRolesResponse, error) {
	if roles, ok := s.adminCache.getRoles(); ok {
		return types.GetRolesResponse{
			Roles:                roles,
			AvailablePermissions: availablePermissions,
		}, nil
	}

	repoRoles, err := s.repo.GetRoles(ctx)
	if err != nil {
		return types.GetRolesResponse{}, err
	}

	roles := make([]types.Role, 0, len(repoRoles))
	for _, role := range repoRoles {
		roles = append(roles, types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: unmaskPermissions(role.Permissions),
		})
	}

	s.adminCache.setRoles(roles)

	return types.GetRolesResponse{
		Roles:                roles,
		AvailablePermissions: availablePermissions,
	}, nil
}

func (s *service) GetRole(ctx context.Context, req types.GetRoleRequest) ([]types.Username, error) {
	if req.RoleID <= 0 {
		return nil, types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	return s.repo.GetRole(ctx, req)
}

func (s *service) UpdateRole(ctx context.Context, req types.UpdateRoleRequest) error {
	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if (req.Name == nil || *req.Name == "") && len(req.Permissions) == 0 {
		return types.ErrEmptyUpdateRequest
	}

	var permissions *uint64
	if len(req.Permissions) > 0 {
		if err := validatePermissions(req.Permissions); err != nil {
			return err
		}

		value := maskPermissions(req.Permissions)
		permissions = &value
	}

	if err := s.repo.UpdateRole(ctx, types.UpdateRoleRepoRequest{
		RoleID:      req.RoleID,
		Name:        req.Name,
		Permissions: permissions,
	}); err != nil {
		return err
	}

	s.adminCache.resetRoles()

	if permissions != nil {
		s.adminCache.resetAllPermissions()
	}

	return nil
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

	if err := s.repo.DeleteRole(ctx, req); err != nil {
		return err
	}

	s.adminCache.resetAllPermissions()
	s.adminCache.resetRoles()

	return nil
}

func (s *service) GetUserPermissions(ctx context.Context, req types.GetUserPermissionsRequest) (uint64, error) {
	if perms, ok := s.adminCache.getPermissions(req.Username); ok {
		return perms, nil
	}

	perms, err := s.repo.GetUserPermissions(ctx, req)
	if err != nil {
		return 0, nil
	}

	s.adminCache.setPermissions(req.Username, perms)

	return perms, nil
}

func (s *service) GetAvailablePermissions() []types.Permission {
	return availablePermissions
}
