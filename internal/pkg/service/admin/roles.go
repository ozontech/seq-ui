package admin

import (
	"context"
	"slices"

	"github.com/ozontech/seq-ui/internal/app/types"
)

func (s *service) CreateRole(ctx context.Context, req types.CreateRoleRequest) (int32, error) {
	if err := s.checkAccess(ctx, permissionCreateRoles); err != nil {
		return 0, err
	}

	if req.Name == "" {
		return 0, types.NewErrInvalidRequestField("empty role name")
	}

	if err := s.validatePermissions(ctx, req.Permissions); err != nil {
		return 0, err
	}

	roleID, err := s.repo.CreateRole(ctx, req)
	if err != nil {
		return 0, err
	}

	s.cache.resetRoles()

	return roleID, nil
}

func (s *service) AddUsersToRole(ctx context.Context, req types.AddUsersToRoleRequest) error {
	if err := s.checkAccess(ctx, permissionUpdateRoles); err != nil {
		return err
	}

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

	s.cache.resetPermissions(req.Usernames...)

	return nil
}

func (s *service) GetRoles(ctx context.Context) (types.GetRolesResponse, error) {
	if err := s.checkAccess(ctx, permissionReadRoles); err != nil {
		return types.GetRolesResponse{}, err
	}

	available, err := s.GetAvailablePermissions(ctx)
	if err != nil {
		return types.GetRolesResponse{}, err
	}

	if roles, ok := s.cache.getRoles(); ok {
		return types.GetRolesResponse{
			Roles:                roles,
			AvailablePermissions: available,
		}, nil
	}

	roles, err := s.repo.GetRoles(ctx)
	if err != nil {
		return types.GetRolesResponse{}, err
	}

	s.cache.setRoles(roles)

	return types.GetRolesResponse{
		Roles:                roles,
		AvailablePermissions: available,
	}, nil
}

func (s *service) GetRole(ctx context.Context, req types.GetRoleRequest) (types.RoleInfo, error) {
	if err := s.checkAccess(ctx, permissionReadRoles); err != nil {
		return types.RoleInfo{}, err
	}

	if req.RoleID <= 0 {
		return types.RoleInfo{}, types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	return s.repo.GetRole(ctx, req)
}

func (s *service) UpdateRole(ctx context.Context, req types.UpdateRoleRequest) error {
	if err := s.checkAccess(ctx, permissionUpdateRoles); err != nil {
		return err
	}

	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if (req.Name == nil || *req.Name == "") && len(req.Permissions) == 0 {
		return types.ErrEmptyUpdateRequest
	}

	if len(req.Permissions) > 0 {
		if err := s.validatePermissions(ctx, req.Permissions); err != nil {
			return err
		}
	}

	if err := s.repo.UpdateRole(ctx, req); err != nil {
		return err
	}

	s.cache.resetRoles()

	if len(req.Permissions) > 0 {
		s.cache.resetAllPermissions()
	}

	return nil
}

func (s *service) DeleteRole(ctx context.Context, req types.DeleteRoleRequest) error {
	if err := s.checkAccess(ctx, permissionDeleteRoles); err != nil {
		return err
	}

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

	s.cache.resetAllPermissions()
	s.cache.resetRoles()

	return nil
}

func (s *service) DeleteUsersFromRole(ctx context.Context, req types.DeleteUsersFromRoleRequest) error {
	if err := s.checkAccess(ctx, permissionDeleteRoles); err != nil {
		return err
	}

	if req.RoleID <= 0 {
		return types.NewErrInvalidRequestField("value role_id must be greater than 0")
	}

	if len(req.Usernames) == 0 {
		return types.NewErrInvalidRequestField("empty usernames")
	}

	if slices.Contains(req.Usernames, "") {
		return types.NewErrInvalidRequestField("empty username")
	}

	if err := s.repo.DeleteUsersFromRole(ctx, req); err != nil {
		return err
	}

	s.cache.resetPermissions(req.Usernames...)

	return nil
}
