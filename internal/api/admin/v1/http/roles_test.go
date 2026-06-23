package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/admin/mock"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
)

func setupTestAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func withRoleID(h http.HandlerFunc, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("id", id)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))
		h(w, r)
	}
}

func TestServeCreateRole(t *testing.T) {
	type mockArgs struct {
		req    types.CreateRoleRequest
		roleID int32
		err    error
	}

	tests := []struct {
		name string

		req     createRoleRequest
		want    createRoleResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: createRoleRequest{
				Name:        "manager",
				Permissions: []string{"roles:create"},
			},
			want: createRoleResponse{RoleID: 1},
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "manager",
					Permissions: []string{"roles:create"},
				},
				roleID: 1,
			},
		},
		{
			name: "err_svc",
			req: createRoleRequest{
				Name:        "manager",
				Permissions: []string{"roles:create"},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "manager",
					Permissions: []string{"roles:create"},
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					CreateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleID, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[createRoleRequest, createRoleResponse]{
				Method:  http.MethodPost,
				Target:  "/admin/v1/roles",
				Req:     tt.req,
				Handler: api.serveCreateRole,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeAddUsersToRole(t *testing.T) {
	type mockArgs struct {
		req types.AddUsersToRoleRequest
		err error
	}

	tests := []struct {
		name string

		roleID  string
		req     addUsersToRoleRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:   "ok",
			roleID: "1",
			req: addUsersToRoleRequest{
				Usernames: []string{"user1", "user2"},
			},
			mockArgs: &mockArgs{
				req: types.AddUsersToRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name:   "err_svc",
			roleID: "1",
			req: addUsersToRoleRequest{
				Usernames: []string{"user1", "user2"},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.AddUsersToRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					AddUsersToRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[addUsersToRoleRequest, struct{}]{
				Method:  http.MethodPost,
				Target:  "/admin/v1/roles/" + tt.roleID + "/users",
				Req:     tt.req,
				Handler: withRoleID(api.serveAddUsersToRole, tt.roleID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeGetRoles(t *testing.T) {
	type mockArgs struct {
		resp types.GetRolesResponse
		err  error
	}

	tests := []struct {
		name string

		want    getRolesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: getRolesResponse{
				Roles: []role{
					{
						ID:          1,
						Name:        "manager",
						Permissions: []string{"roles:delete"},
					},
				},
				AvailablePermissions: []permission{
					{ID: 1, Value: "roles:create"},
					{ID: 2, Value: "roles:read"},
					{ID: 3, Value: "roles:update"},
					{ID: 4, Value: "roles:delete"},
				},
			},
			mockArgs: &mockArgs{
				resp: types.GetRolesResponse{
					Roles: []types.Role{
						{
							ID:          1,
							Name:        "manager",
							Permissions: []string{"roles:delete"},
						},
					},
					AvailablePermissions: []types.Permission{
						{ID: 1, Value: "roles:create"},
						{ID: 2, Value: "roles:read"},
						{ID: 3, Value: "roles:update"},
						{ID: 4, Value: "roles:delete"},
					},
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetRoles(gomock.Any()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getRolesResponse]{
				Method:  http.MethodGet,
				Target:  "/admin/v1/roles",
				Handler: api.serveGetRoles,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeGetRole(t *testing.T) {
	type mockArgs struct {
		req  types.GetRoleRequest
		resp types.RoleInfo
		err  error
	}

	tests := []struct {
		name string

		roleID  string
		want    getRoleResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:   "ok",
			roleID: "1",
			want: getRoleResponse{
				Usernames: []string{"user1", "user2"},
			},
			mockArgs: &mockArgs{
				req: types.GetRoleRequest{RoleID: 1},
				resp: types.RoleInfo{
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name:    "err_svc",
			roleID:  "1",
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.GetRoleRequest{RoleID: 1},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getRoleResponse]{
				Method:  http.MethodGet,
				Target:  "/admin/v1/roles/" + tt.roleID,
				Handler: withRoleID(api.serveGetRole, tt.roleID),
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeUpdateRole(t *testing.T) {
	name := "new role name"

	type mockArgs struct {
		req types.UpdateRoleRequest
		err error
	}

	tests := []struct {
		name string

		roleID  string
		req     updateRoleRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:   "ok",
			roleID: "1",
			req: updateRoleRequest{
				Name:        &name,
				Permissions: []string{"roles:delete"},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: []string{"roles:delete"},
				},
			},
		},
		{
			name:   "err_svc",
			roleID: "1",
			req: updateRoleRequest{
				Name:        &name,
				Permissions: []string{"roles:delete"},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: []string{"roles:delete"},
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					UpdateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[updateRoleRequest, struct{}]{
				Method:  http.MethodPatch,
				Target:  "/admin/v1/roles/" + tt.roleID,
				Req:     tt.req,
				Handler: withRoleID(api.serveUpdateRole, tt.roleID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeDeleteRole(t *testing.T) {
	replacementID := int32(2)

	type mockArgs struct {
		req types.DeleteRoleRequest
		err error
	}

	tests := []struct {
		name string

		roleID  string
		req     deleteRoleRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:   "ok_no_replacement",
			roleID: "1",
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{RoleID: 1},
			},
		},
		{
			name:   "ok_with_replacement",
			roleID: "1",
			req: deleteRoleRequest{
				ReplacementRoleID: &replacementID,
			},
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{
					RoleID:            1,
					ReplacementRoleID: &replacementID,
				},
			},
		},
		{
			name:    "err_svc",
			roleID:  "1",
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{RoleID: 1},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					DeleteRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[deleteRoleRequest, struct{}]{
				Method:  http.MethodDelete,
				Target:  "/admin/v1/roles/" + tt.roleID,
				Req:     tt.req,
				Handler: withRoleID(api.serveDeleteRole, tt.roleID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeDeleteUsersFromRole(t *testing.T) {
	type mockArgs struct {
		req types.DeleteUsersFromRoleRequest
		err error
	}

	tests := []struct {
		name string

		roleID  string
		req     deleteUsersFromRoleRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:   "ok",
			roleID: "1",
			req: deleteUsersFromRoleRequest{
				Usernames: []string{"user1", "user2"},
			},
			mockArgs: &mockArgs{
				req: types.DeleteUsersFromRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name:   "err_svc",
			roleID: "1",
			req: deleteUsersFromRoleRequest{
				Usernames: []string{"user1"},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.DeleteUsersFromRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1"},
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					DeleteUsersFromRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[deleteUsersFromRoleRequest, struct{}]{
				Method:  http.MethodDelete,
				Target:  "/admin/v1/roles/" + tt.roleID + "/users",
				Req:     tt.req,
				Handler: withRoleID(api.serveDeleteUsersFromRole, tt.roleID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}
