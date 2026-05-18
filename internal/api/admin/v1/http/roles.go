package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) serveCreateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_create_role")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_name",
			Value: attribute.StringValue(httpReq.Name),
		},
		attribute.KeyValue{
			Key:   "permissions_count",
			Value: attribute.IntValue(len(httpReq.Permissions)),
		},
	)

	roleID, err := a.service.CreateRole(ctx, types.CreateRoleRequest{
		Name:        httpReq.Name,
		Permissions: httpReq.Permissions,
	})
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(createRoleResponse{RoleID: roleID})
}

func (a *API) serveAddUsersToRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_add_users_to_role")
	defer span.End()

	wr := httputil.NewWriter(w)

	roleID, err := getRoleID(r)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	var httpReq addUsersToRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_id",
			Value: attribute.IntValue(int(roleID)),
		},
		attribute.KeyValue{
			Key:   "users_count",
			Value: attribute.IntValue(len(httpReq.Usernames)),
		},
	)

	if err := a.service.AddUsersToRole(ctx, types.AddUsersToRoleRequest{
		RoleID:    roleID,
		Usernames: httpReq.Usernames,
	}); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) serveGetRoles(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_get_roles")
	defer span.End()

	wr := httputil.NewWriter(w)

	resp, err := a.service.GetRoles(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getRolesResponse{
		Roles:                parseRoles(resp.Roles),
		AvailablePermissions: a.availablePermissions,
	})
}

func (a *API) serveGetRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_get_role")
	defer span.End()

	wr := httputil.NewWriter(w)

	roleID, err := getRoleID(r)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "role_id",
			Value: attribute.IntValue(int(roleID)),
		},
	)

	usernames, err := a.service.GetRole(ctx, types.GetRoleRequest{
		RoleID: roleID,
	})
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getRoleResponse{
		Usernames: usernames,
	})
}

func (a *API) serveUpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_update_role")
	defer span.End()

	wr := httputil.NewWriter(w)

	roleID, err := getRoleID(r)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	var httpReq updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "role_id",
			Value: attribute.IntValue(int(roleID)),
		},
		{
			Key:   "permissions_count",
			Value: attribute.IntValue(len(httpReq.Permissions)),
		},
	}
	if httpReq.Name != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "role_name",
			Value: attribute.StringValue(*httpReq.Name),
		})
	}
	span.SetAttributes(spanAttributes...)

	if err := a.service.UpdateRole(ctx, types.UpdateRoleRequest{
		RoleID:      roleID,
		Name:        httpReq.Name,
		Permissions: httpReq.Permissions,
	}); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) serveDeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "admin_v1_delete_role")
	defer span.End()

	wr := httputil.NewWriter(w)

	roleID, err := getRoleID(r)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	var httpReq deleteRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil && !errors.Is(err, io.EOF) {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "role_id",
			Value: attribute.IntValue(int(roleID)),
		},
	}
	if httpReq.ReplacementRoleID != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "replacement_role_id",
			Value: attribute.IntValue(int(*httpReq.ReplacementRoleID)),
		})
	}
	span.SetAttributes(spanAttributes...)

	if err := a.service.DeleteRole(ctx, types.DeleteRoleRequest{
		RoleID:            roleID,
		ReplacementRoleID: httpReq.ReplacementRoleID,
	}); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getRoleID(r *http.Request) (int32, error) {
	idString := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idString, 10, 32)
	if err != nil {
		return 0, types.NewErrInvalidRequestField("invalid role_id")
	}

	return int32(id), nil
}

func parseRoles(source []types.Role) []role {
	roles := make([]role, 0, len(source))
	for _, s := range source {
		roles = append(roles, role{
			ID:          s.ID,
			Name:        s.Name,
			Permissions: s.Permissions,
		})
	}
	return roles
}

func parsePermissions(source []types.Permission) []permission {
	permissions := make([]permission, 0, len(source))
	for _, s := range source {
		permissions = append(permissions, permission{
			Value:       s.Value,
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return permissions
}

type role struct {
	ID          int32    `json:"id"`
	Name        string   `json:"name"`
	Permissions []uint64 `json:"permissions"`
}

type permission struct {
	Value       uint64 `json:"value"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []uint64 `json:"permissions"`
}

type createRoleResponse struct {
	RoleID int32 `json:"role_id"`
}

type addUsersToRoleRequest struct {
	Usernames []string `json:"usernames"`
}

type getRolesResponse struct {
	Roles                []role       `json:"roles"`
	AvailablePermissions []permission `json:"available_permissions"`
}

type getRoleResponse struct {
	Usernames []string `json:"usernames"`
}

type updateRoleRequest struct {
	Name        *string  `json:"name"`
	Permissions []uint64 `json:"permissions"`
}

type deleteRoleRequest struct {
	ReplacementRoleID *int32 `json:"replacement_role_id"`
}
