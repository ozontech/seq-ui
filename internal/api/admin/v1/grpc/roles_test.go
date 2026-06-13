package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/admin/mock"
	"github.com/ozontech/seq-ui/pkg/admin/v1"
)

var (
	errSomethingWrong = errors.New("something happened wrong")
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	return New(mockedSvc), mockedSvc
}

func TestCreateRole(t *testing.T) {
	type mockArgs struct {
		req    types.CreateRoleRequest
		roleID int32
		err    error
	}

	tests := []struct {
		name string

		req      *admin.CreateRoleRequest
		want     *admin.CreateRoleResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &admin.CreateRoleRequest{
				Name:        "manager",
				Permissions: []uint64{1},
			},
			want:     &admin.CreateRoleResponse{RoleId: 1},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "manager",
					Permissions: []uint64{1},
				},
				roleID: 1,
			},
		},
		{
			name: "err_svc",
			req: &admin.CreateRoleRequest{
				Name:        "manager",
				Permissions: []uint64{1},
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "manager",
					Permissions: []uint64{1},
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					CreateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleID, tt.mockArgs.err).
					Times(1)
			}

			gotResp, err := api.CreateRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, gotResp)
		})
	}
}

func TestAddUsersToRole(t *testing.T) {
	type mockArgs struct {
		req types.AddUsersToRoleRequest
		err error
	}

	tests := []struct {
		name string

		req      *admin.AddUsersToRoleRequest
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &admin.AddUsersToRoleRequest{
				RoleId:    1,
				Usernames: []string{"user1", "user2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.AddUsersToRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name: "err_svc",
			req: &admin.AddUsersToRoleRequest{
				RoleId:    1,
				Usernames: []string{"user1"},
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.AddUsersToRoleRequest{
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

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					AddUsersToRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			_, err := api.AddUsersToRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestGetRoles(t *testing.T) {
	type mockArgs struct {
		resp types.GetRolesResponse
		err  error
	}

	tests := []struct {
		name string

		want     *admin.GetRolesResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: &admin.GetRolesResponse{
				Roles: []*admin.Role{
					{Id: 1, Name: "manager", Permissions: []uint64{1}},
				},
				AvailablePermissions: []*admin.GetRolesResponse_Permission{
					{Value: 1, Name: "manage_roles", Description: "Manage roles"},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				resp: types.GetRolesResponse{
					Roles: []types.Role{
						{ID: 1, Name: "manager", Permissions: []uint64{1}},
					},
				},
			},
		},
		{
			name:     "err_svc",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetRoles(gomock.Any()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			gotResp, err := api.GetRoles(context.Background(), &admin.GetRolesRequest{})
			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, gotResp)
		})
	}
}

func TestGetRole(t *testing.T) {
	type mockArgs struct {
		req      types.GetRoleRequest
		roleInfo types.RoleInfo
		err      error
	}

	tests := []struct {
		name string

		req      *admin.GetRoleRequest
		want     *admin.GetRoleResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  &admin.GetRoleRequest{Id: 1},
			want: &admin.GetRoleResponse{
				Usernames: []string{"user1", "user2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req:      types.GetRoleRequest{RoleID: 1},
				roleInfo: types.RoleInfo{Usernames: []string{"user1", "user2"}},
			},
		},
		{
			name:     "err_svc",
			req:      &admin.GetRoleRequest{Id: 1},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetRoleRequest{RoleID: 1},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleInfo, tt.mockArgs.err).
					Times(1)
			}

			gotResp, err := api.GetRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, gotResp)
		})
	}
}

func TestUpdateRole(t *testing.T) {
	name := "new role name"

	type mockArgs struct {
		req types.UpdateRoleRequest
		err error
	}

	tests := []struct {
		name string

		req      *admin.UpdateRoleRequest
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &admin.UpdateRoleRequest{
				Id:          1,
				Name:        &name,
				Permissions: []uint64{1},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: []uint64{1},
				},
			},
		},
		{
			name: "err_svc",
			req: &admin.UpdateRoleRequest{
				Id:   1,
				Name: &name,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID: 1,
					Name:   &name,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					UpdateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			_, err := api.UpdateRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestDeleteRole(t *testing.T) {
	replacementID := int32(2)

	type mockArgs struct {
		req types.DeleteRoleRequest
		err error
	}

	tests := []struct {
		name string

		req      *admin.DeleteRoleRequest
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name:     "ok_no_replacement",
			req:      &admin.DeleteRoleRequest{Id: 1},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{RoleID: 1},
			},
		},
		{
			name: "ok_with_replacement",
			req: &admin.DeleteRoleRequest{
				Id:                1,
				ReplacementRoleId: &replacementID,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{
					RoleID:            1,
					ReplacementRoleID: &replacementID,
				},
			},
		},
		{
			name:     "err_svc",
			req:      &admin.DeleteRoleRequest{Id: 1},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{RoleID: 1},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					DeleteRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			_, err := api.DeleteRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestDeleteUsersFromRole(t *testing.T) {
	type mockArgs struct {
		req types.DeleteUsersFromRoleRequest
		err error
	}

	tests := []struct {
		name string

		req      *admin.DeleteUsersFromRoleRequest
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &admin.DeleteUsersFromRoleRequest{
				RoleId:    1,
				Usernames: []string{"user1", "user2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteUsersFromRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name: "err_svc",
			req: &admin.DeleteUsersFromRoleRequest{
				RoleId:    1,
				Usernames: []string{"user1"},
			},
			wantCode: codes.Internal,
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

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					DeleteUsersFromRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			_, err := api.DeleteUsersFromRole(context.Background(), tt.req)
			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
