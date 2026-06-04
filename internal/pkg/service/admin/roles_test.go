package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	defaultSuperUser  = "superuser"
	adminCfg          = &config.Admin{SuperUsers: []string{defaultSuperUser}}
	errSomethingWrong = errors.New("something happened wrong")
)

type accessMock struct {
	permissions uint64
	err         error
}

func setupAccessMock(repo *mock.MockAdmin, actorUsername string, access *accessMock) {
	if access == nil {
		return
	}

	repo.EXPECT().
		GetUserPermissions(gomock.Any(), types.GetUserPermissionsRequest{Username: actorUsername}).
		Return(access.permissions, access.err).
		Times(1)
}

func TestCreateRole(t *testing.T) {
	type MockArgs struct {
		req    types.CreateRoleRepoRequest
		roleID int32
		err    error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.CreateRoleRequest

		accessMock *accessMock
		mockArgs   *MockArgs

		wantRoleID int32
		wantErr    bool
	}{
		{
			name:          "ok",
			actorUsername: "admin",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{permissionManageRoles},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &MockArgs{
				req: types.CreateRoleRepoRequest{
					Name:        "typical good boy",
					Permissions: permissionManageRoles,
				},
				roleID: 1,
			},
			wantRoleID: 1,
		},
		{
			name:          "ok_superuser",
			actorUsername: defaultSuperUser,
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{permissionManageRoles},
			},
			mockArgs: &MockArgs{
				req: types.CreateRoleRepoRequest{
					Name:        "typical good boy",
					Permissions: permissionManageRoles,
				},
				roleID: 2,
			},
			wantRoleID: 2,
		},
		{
			name: "err_no_auth",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{permissionManageRoles},
			},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "bad boy",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{permissionManageRoles},
			},
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_name",
			actorUsername: "admin",
			req: types.CreateRoleRequest{
				Name:        "",
				Permissions: []uint64{permissionManageRoles},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_permissions",
			actorUsername: "admin",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_unknown_permissions",
			actorUsername: "admin",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{52},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []uint64{permissionManageRoles},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &MockArgs{
				req: types.CreateRoleRepoRequest{
					Name:        "typical good boy",
					Permissions: permissionManageRoles,
				},
				err: errSomethingWrong,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			ctx := context.Background()

			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					CreateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleID, tt.mockArgs.err).
					Times(1)
			}

			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			roleID, err := svc.CreateRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRoleID, roleID)
		})
	}
}

func TestAddUsersToRole(t *testing.T) {
	type mockArgs struct {
		err error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.AddUsersToRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name:          "ok",
			actorUsername: "admin",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", "user2"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{},
		},
		{
			name:          "ok_superuser",
			actorUsername: defaultSuperUser,
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			mockArgs: &mockArgs{},
		},
		{
			name: "err_no_auth",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "user",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_invalid_role_id_zero",
			actorUsername: "admin",
			req: types.AddUsersToRoleRequest{
				RoleID:    0,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_usernames",
			actorUsername: "admin",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_contains_empty_username",
			actorUsername: "admin",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", ""},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			ctx := context.Background()

			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					AddUsersToRole(gomock.Any(), tt.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			err := svc.AddUsersToRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetRoles(t *testing.T) {
	type mockArgs struct {
		roles []types.RoleRepo
		err   error
	}

	tests := []struct {
		name          string
		actorUsername string

		accessMock *accessMock
		mockArgs   *mockArgs

		wantRoles []types.Role
		wantErr   bool
	}{
		{
			name:          "ok",
			actorUsername: "admin",
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				roles: []types.RoleRepo{
					{
						ID:          1,
						Name:        "admin",
						Permissions: permissionManageRoles,
					},
				},
			},
			wantRoles: []types.Role{
				{
					ID:          1,
					Name:        "admin",
					Permissions: []uint64{permissionManageRoles},
				},
			},
		},
		{
			name:          "ok_empty",
			actorUsername: "admin",
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				roles: []types.RoleRepo{},
			},
			wantRoles: []types.Role{},
		},
		{
			name:    "err_no_auth",
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "user",
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetRoles(gomock.Any()).
					Return(tt.mockArgs.roles, tt.mockArgs.err).
					Times(1)
			}

			ctx := context.Background()
			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			// first call goes to repo.
			respFromRepo, err := svc.GetRoles(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRoles, respFromRepo.Roles)
			require.Equal(t, availablePermissions, respFromRepo.AvailablePermissions)

			// second call served from cache.
			respFromCache, err := svc.GetRoles(ctx)
			require.NoError(t, err)
			require.Equal(t, respFromRepo, respFromCache)
		})
	}
}

func TestGetRole(t *testing.T) {
	type mockArgs struct {
		roleInfo types.RoleInfo
		err      error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.GetRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		want    types.RoleInfo
		wantErr bool
	}{
		{
			name:          "ok",
			actorUsername: "admin",
			req:           types.GetRoleRequest{RoleID: 1},
			want: types.RoleInfo{
				Usernames: []string{"user1", "user2"},
			},
			mockArgs: &mockArgs{
				roleInfo: types.RoleInfo{
					Usernames: []string{"user1", "user2"},
				},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
		},
		{
			name:    "err_no_auth",
			req:     types.GetRoleRequest{RoleID: 1},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "user",
			req:           types.GetRoleRequest{RoleID: 1},
			wantErr:       true,
			accessMock: &accessMock{
				permissions: 0,
			},
		},
		{
			name:          "err_invalid_role_id_zero",
			actorUsername: "admin",
			req:           types.GetRoleRequest{RoleID: 0},
			wantErr:       true,
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			req:           types.GetRoleRequest{RoleID: 1},
			wantErr:       true,
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetRole(gomock.Any(), tt.req).
					Return(tt.mockArgs.roleInfo, tt.mockArgs.err).
					Times(1)
			}

			ctx := context.Background()
			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			gotRoleInfo, err := svc.GetRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, gotRoleInfo)
		})
	}
}

func TestUpdateRole(t *testing.T) {
	name := "new role name"
	emptyName := ""

	type mockArgs struct {
		req types.UpdateRoleRepoRequest
		err error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.UpdateRoleRequest

		mockArgs   *mockArgs
		accessMock *accessMock

		wantErr bool
	}{
		{
			name:          "ok_name_and_permissions",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Name:        &name,
				Permissions: []uint64{permissionManageRoles},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRepoRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: ptr(permissionManageRoles),
				},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
		},
		{
			name:          "ok_name_only",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRepoRequest{
					RoleID: 1,
					Name:   &name,
				},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
		},
		{
			name:          "ok_permissions_only",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []uint64{permissionManageRoles},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRepoRequest{
					RoleID:      1,
					Permissions: ptr(permissionManageRoles),
				},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
		},
		{
			name:          "ok_superuser",
			actorUsername: defaultSuperUser,
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRepoRequest{
					RoleID: 1,
					Name:   &name,
				},
			},
		},
		{
			name: "err_no_auth",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "typical bad boy",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_invalid_role_id",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID: 0,
				Name:   &name,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_update",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_name_no_permissions",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &emptyName,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_unknown_permission",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []uint64{52},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "typical good boy",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRepoRequest{
					RoleID: 1,
					Name:   &name,
				},
				err: errSomethingWrong,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					UpdateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			ctx := context.Background()
			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			err := svc.UpdateRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDeleteRole(t *testing.T) {
	replacementID := int32(2)

	type mockArgs struct {
		repoReq types.DeleteRoleRequest
		errRepo error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.DeleteRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name:          "ok_no_replacement",
			actorUsername: "admin",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				repoReq: types.DeleteRoleRequest{RoleID: 1},
			},
		},
		{
			name:          "ok_with_replacement",
			actorUsername: "admin",
			req: types.DeleteRoleRequest{
				RoleID:            1,
				ReplacementRoleID: &replacementID,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				repoReq: types.DeleteRoleRequest{
					RoleID:            1,
					ReplacementRoleID: &replacementID,
				},
			},
		},
		{
			name: "err_no_auth",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "user",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_invalid_role_id",
			actorUsername: "admin",
			req: types.DeleteRoleRequest{
				RoleID: 0,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_replacement_equals_role_id",
			actorUsername: "admin",
			req: types.DeleteRoleRequest{
				RoleID:            1,
				ReplacementRoleID: ptr[int32](1),
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				repoReq: types.DeleteRoleRequest{RoleID: 1},
				errRepo: errSomethingWrong,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)
			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					DeleteRole(gomock.Any(), tt.mockArgs.repoReq).
					Return(tt.mockArgs.errRepo).
					Times(1)
			}

			ctx := context.Background()
			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			err := svc.DeleteRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDeleteUsersFromRole(t *testing.T) {
	type mockArgs struct {
		req types.DeleteUsersFromRoleRequest
		err error
	}

	tests := []struct {
		name          string
		actorUsername string
		req           types.DeleteUsersFromRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name:          "ok",
			actorUsername: "admin",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", "user2"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				req: types.DeleteUsersFromRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1", "user2"},
				},
			},
		},
		{
			name: "err_no_auth",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"u"},
			},
			wantErr: true,
		},
		{
			name:          "err_no_access",
			actorUsername: "user",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: 0,
			},
			wantErr: true,
		},
		{
			name:          "err_invalid_role_id",
			actorUsername: "admin",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    0,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_empty_usernames",
			actorUsername: "admin",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_contains_empty_username",
			actorUsername: "admin",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", ""},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			wantErr: true,
		},
		{
			name:          "err_repo",
			actorUsername: "admin",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				permissions: permissionManageRoles,
			},
			mockArgs: &mockArgs{
				req: types.DeleteUsersFromRoleRequest{
					RoleID:    1,
					Usernames: []string{"user1"},
				},
				err: errSomethingWrong,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg)

			setupAccessMock(repo, tt.actorUsername, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					DeleteUsersFromRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			ctx := context.Background()
			if tt.actorUsername != "" {
				ctx = types.SetUserKey(ctx, tt.actorUsername)
			}

			err := svc.DeleteUsersFromRole(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}
