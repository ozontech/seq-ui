package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) CreateRole(ctx context.Context, req *admin.CreateRoleRequest) (*admin.CreateRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_create_role")
	defer span.End()

	request := types.CreateRoleRequest{
		Name:        req.Name,
		Permissions: req.Permissions,
	}

	roleID, err := a.service.CreateRole(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.CreateRoleResponse{
		RoleId: roleID.RoleID,
	}, nil
}

func (a *API) AddUsersToRole(ctx context.Context, req *admin.AddUsersToRoleRequest) (*admin.AddUsersToRoleResponse, error) {
	return &admin.AddUsersToRoleResponse{}, nil
}

func (a *API) GetRoles(ctx context.Context, req *admin.GetRolesRequest) (*admin.GetRolesResponse, error) {
	return &admin.GetRolesResponse{}, nil
}

func (a *API) GetRole(ctx context.Context, req *admin.GetRoleRequest) (*admin.GetRoleResponse, error) {
	return &admin.GetRoleResponse{}, nil
}

func (a *API) UpdateRole(ctx context.Context, req *admin.UpdateRoleRequest) (*admin.UpdateRoleResponse, error) {
	return &admin.UpdateRoleResponse{}, nil
}

func (a *API) DeleteRole(ctx context.Context, req *admin.DeleteRoleRequest) (*admin.DeleteRoleResponse, error) {
	return &admin.DeleteRoleResponse{}, nil
}
