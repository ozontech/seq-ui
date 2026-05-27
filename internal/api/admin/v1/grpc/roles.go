package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) CreateRole(ctx context.Context, req *admin.CreateRoleRequest) (*admin.CreateRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_create_role")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_name",
			Value: attribute.StringValue(req.GetName()),
		},
		attribute.KeyValue{
			Key:   "permissions_count",
			Value: attribute.IntValue(len(req.GetPermissions())),
		},
	)

	request := types.CreateRoleRequest{
		Name:        req.Name,
		Permissions: req.Permissions,
	}

	roleID, err := a.service.CreateRole(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.CreateRoleResponse{
		RoleId: roleID,
	}, nil
}

func (a *API) AddUsersToRole(ctx context.Context, req *admin.AddUsersToRoleRequest) (*admin.AddUsersToRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_add_users_to_role")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_id",
			Value: attribute.IntValue(int(req.GetRoleId())),
		},
		attribute.KeyValue{
			Key:   "users_count",
			Value: attribute.IntValue(len(req.GetUsernames())),
		},
	)

	if err := a.service.AddUsersToRole(ctx, types.AddUsersToRoleRequest{
		RoleID:    req.RoleId,
		Usernames: req.Usernames,
	}); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.AddUsersToRoleResponse{}, nil
}

func (a *API) GetRoles(ctx context.Context, _ *admin.GetRolesRequest) (*admin.GetRolesResponse, error) {
	resp, err := a.service.GetRoles(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.GetRolesResponse{
		Roles:                rolesToProto(resp.Roles),
		AvailablePermissions: a.availablePermissions,
	}, nil
}

func (a *API) GetRole(ctx context.Context, req *admin.GetRoleRequest) (*admin.GetRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_get_role")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_id",
			Value: attribute.IntValue(int(req.GetId())),
		},
	)

	roleInfo, err := a.service.GetRole(ctx, types.GetRoleRequest{
		RoleID: req.Id,
	})
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.GetRoleResponse{
		Usernames: roleInfo.Usernames,
	}, nil
}

func (a *API) UpdateRole(ctx context.Context, req *admin.UpdateRoleRequest) (*admin.UpdateRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_update_role")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "role_id",
			Value: attribute.IntValue(int(req.GetId())),
		},
		{
			Key:   "permissions_count",
			Value: attribute.IntValue(len(req.GetPermissions())),
		},
	}
	if req.Name != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "role_name",
			Value: attribute.StringValue(req.GetName()),
		})
	}
	span.SetAttributes(spanAttributes...)

	if err := a.service.UpdateRole(ctx, types.UpdateRoleRequest{
		RoleID:      req.Id,
		Name:        req.Name,
		Permissions: req.Permissions,
	}); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.UpdateRoleResponse{}, nil
}

func (a *API) DeleteRole(ctx context.Context, req *admin.DeleteRoleRequest) (*admin.DeleteRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_delete_role")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "role_id",
			Value: attribute.IntValue(int(req.GetId())),
		},
	}
	if req.ReplacementRoleId != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "replacement_role_id",
			Value: attribute.IntValue(int(req.GetReplacementRoleId())),
		})
	}
	span.SetAttributes(spanAttributes...)

	if err := a.service.DeleteRole(ctx, types.DeleteRoleRequest{
		RoleID:            req.Id,
		ReplacementRoleID: req.ReplacementRoleId,
	}); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.DeleteRoleResponse{}, nil
}

func (a *API) DeleteUsersFromRole(ctx context.Context, req *admin.DeleteUsersFromRoleRequest) (*admin.DeleteUsersFromRoleResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "admin_v1_delete_users_from_role")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_id",
			Value: attribute.IntValue(int(req.GetRoleId())),
		},
		attribute.KeyValue{
			Key:   "users_count",
			Value: attribute.IntValue(len(req.GetUsernames())),
		},
	)

	if err := a.service.DeleteUsersFromRole(ctx, types.DeleteUsersFromRoleRequest{
		RoleID:    req.RoleId,
		Usernames: req.Usernames,
	}); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &admin.DeleteUsersFromRoleResponse{}, nil
}

func rolesToProto(source []types.Role) []*admin.Role {
	roles := make([]*admin.Role, 0, len(source))
	for _, role := range source {
		roles = append(roles, &admin.Role{
			Id:          role.ID,
			Name:        role.Name,
			Permissions: role.Permissions,
		})
	}
	return roles
}
