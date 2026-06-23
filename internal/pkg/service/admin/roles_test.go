package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

var (
	defaultSuperUser  = "superuser"
	adminCfg          = &config.Admin{SuperUsers: []string{defaultSuperUser}}
	errSomethingWrong = errors.New("something happened wrong")

	testAvailablePermissions = []types.Permission{
		{ID: 1, Value: "roles:create"},
		{ID: 2, Value: "roles:read"},
		{ID: 3, Value: "roles:update"},
		{ID: 4, Value: "roles:delete"},
	}
)

type accessMock struct {
	username    string
	permissions []string
	err         error
}

func setupAccessMock(ctx context.Context, repo *mock.MockAdmin, access *accessMock) context.Context {
	if access == nil {
		return ctx
	}

	ctx = types.SetUserKey(ctx, access.username)

	if access.username != defaultSuperUser {
		repo.EXPECT().
			GetUserPermissions(gomock.Any(), types.GetUserPermissionsRequest{Username: access.username}).
			Return(access.permissions, access.err).
			Times(1)
	}

	return ctx
}

func TestCreateRole(t *testing.T) {
	type mockArgs struct {
		req    types.CreateRoleRequest
		roleID int32
		err    error
	}

	tests := []struct {
		name string
		req  types.CreateRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		needValidateMock bool
		wantRoleID       int32
		wantErr          bool
	}{
		{
			name: "ok",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionCreateRoles},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionCreateRoles},
			},
			needValidateMock: true,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionCreateRoles},
				},
				roleID: 1,
			},
			wantRoleID: 1,
		},
		{
			name: "ok_superuser",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionCreateRoles},
			},
			needValidateMock: true,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionCreateRoles},
				},
				roleID: 2,
			},
			accessMock: &accessMock{
				username: defaultSuperUser,
			},
			wantRoleID: 2,
		},
		{
			name: "err_no_auth",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionCreateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_no_access",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionCreateRoles},
			},
			accessMock: &accessMock{
				username:    "bad boy",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_empty_name",
			req: types.CreateRoleRequest{
				Name:        "",
				Permissions: []string{permissionCreateRoles},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionCreateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_empty_permissions",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionCreateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_unknown_permissions",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{"unknown:operation"},
			},
			needValidateMock: true,
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionCreateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionCreateRoles},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionCreateRoles},
			},
			needValidateMock: true,
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionCreateRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.needValidateMock {
				repo.EXPECT().
					GetAvailablePermissions(gomock.Any()).
					Return(testAvailablePermissions, nil).
					Times(1)
			}

			if tt.mockArgs != nil {
				repo.EXPECT().
					CreateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleID, tt.mockArgs.err).
					Times(1)
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
		name string
		req  types.AddUsersToRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name: "ok",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", "user2"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionUpdateRoles},
			},
			mockArgs: &mockArgs{},
		},
		{
			name: "ok_superuser",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			mockArgs: &mockArgs{},
			accessMock: &accessMock{
				username: defaultSuperUser,
			},
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
			name: "err_no_access",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "user",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_invalid_role_id_zero",
			req: types.AddUsersToRoleRequest{
				RoleID:    0,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_empty_usernames",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_contains_empty_username",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", ""},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.AddUsersToRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionUpdateRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					AddUsersToRole(gomock.Any(), tt.req).
					Return(tt.mockArgs.err).
					Times(1)
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
		roles []types.Role
		err   error
	}

	tests := []struct {
		name       string
		accessMock *accessMock
		mockArgs   *mockArgs

		wantRoles []types.Role
		wantErr   bool
	}{
		{
			name: "ok",
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
			},
			mockArgs: &mockArgs{
				roles: []types.Role{
					{
						ID:          1,
						Name:        "admin",
						Permissions: []string{permissionCreateRoles, permissionUpdateRoles},
					},
				},
			},
			wantRoles: []types.Role{
				{
					ID:          1,
					Name:        "admin",
					Permissions: []string{permissionCreateRoles, permissionUpdateRoles},
				},
			},
		},
		{
			name: "ok_empty",
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
			},
			mockArgs: &mockArgs{
				roles: []types.Role{},
			},
			wantRoles: []types.Role{},
		},
		{
			name:    "err_no_auth",
			wantErr: true,
		},
		{
			name: "err_no_access",
			accessMock: &accessMock{
				username:    "user",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetAvailablePermissions(gomock.Any()).
					Return(testAvailablePermissions, nil).
					Times(1)

				repo.EXPECT().
					GetRoles(gomock.Any()).
					Return(tt.mockArgs.roles, tt.mockArgs.err).
					Times(1)
			}

			// first call goes to repo.
			respFromRepo, err := svc.GetRoles(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRoles, respFromRepo.Roles)
			require.Equal(t, testAvailablePermissions, respFromRepo.AvailablePermissions)

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
		name string
		req  types.GetRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		want    types.RoleInfo
		wantErr bool
	}{
		{
			name: "ok",
			req:  types.GetRoleRequest{RoleID: 1},
			want: types.RoleInfo{
				Usernames: []string{"user1", "user2"},
			},
			mockArgs: &mockArgs{
				roleInfo: types.RoleInfo{
					Usernames: []string{"user1", "user2"},
				},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
			},
		},
		{
			name:    "err_no_auth",
			req:     types.GetRoleRequest{RoleID: 1},
			wantErr: true,
		},
		{
			name:    "err_no_access",
			req:     types.GetRoleRequest{RoleID: 1},
			wantErr: true,
			accessMock: &accessMock{
				username:    "user",
				permissions: []string{},
			},
		},
		{
			name:    "err_invalid_role_id_zero",
			req:     types.GetRoleRequest{RoleID: 0},
			wantErr: true,
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
			},
		},
		{
			name:    "err_repo",
			req:     types.GetRoleRequest{RoleID: 1},
			wantErr: true,
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionReadRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetRole(gomock.Any(), tt.req).
					Return(tt.mockArgs.roleInfo, tt.mockArgs.err).
					Times(1)
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
		req types.UpdateRoleRequest
		err error
	}

	tests := []struct {
		name string
		req  types.UpdateRoleRequest

		mockArgs   *mockArgs
		accessMock *accessMock

		needValidateMock bool
		wantErr          bool
	}{
		{
			name: "ok_name_and_permissions",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Name:        &name,
				Permissions: []string{permissionReadRoles},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: []string{permissionReadRoles},
				},
			},
			needValidateMock: true,
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
		},
		{
			name: "ok_name_only",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID: 1,
					Name:   &name,
				},
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
		},
		{
			name: "ok_permissions_only",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []string{permissionReadRoles},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Permissions: []string{permissionReadRoles},
				},
			},
			needValidateMock: true,
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
		},
		{
			name: "ok_superuser",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID: 1,
					Name:   &name,
				},
			},
			accessMock: &accessMock{
				username: defaultSuperUser,
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
			name: "err_no_access",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			accessMock: &accessMock{
				username:    "typical bad boy",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_invalid_role_id",
			req: types.UpdateRoleRequest{
				RoleID: 0,
				Name:   &name,
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_empty_update",
			req: types.UpdateRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_empty_name_no_permissions",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &emptyName,
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_unknown_permission",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []string{"5:2"},
			},
			needValidateMock: true,
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.UpdateRoleRequest{
				RoleID: 1,
				Name:   &name,
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID: 1,
					Name:   &name,
				},
				err: errSomethingWrong,
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionUpdateRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.needValidateMock {
				repo.EXPECT().
					GetAvailablePermissions(gomock.Any()).
					Return(testAvailablePermissions, nil).
					Times(1)
			}

			if tt.mockArgs != nil {
				repo.EXPECT().
					UpdateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
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
		name string
		req  types.DeleteRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name: "ok_no_replacement",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			mockArgs: &mockArgs{
				repoReq: types.DeleteRoleRequest{RoleID: 1},
			},
		},
		{
			name: "ok_with_replacement",
			req: types.DeleteRoleRequest{
				RoleID:            1,
				ReplacementRoleID: &replacementID,
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
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
			name: "err_no_access",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				username:    "user",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_invalid_role_id",
			req: types.DeleteRoleRequest{
				RoleID: 0,
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			wantErr: true,
		},
		{
			name: "err_replacement_equals_role_id",
			req: types.DeleteRoleRequest{
				RoleID:            1,
				ReplacementRoleID: ptr[int32](1),
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.DeleteRoleRequest{
				RoleID: 1,
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					DeleteRole(gomock.Any(), tt.mockArgs.repoReq).
					Return(tt.mockArgs.errRepo).
					Times(1)
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
		name string
		req  types.DeleteUsersFromRoleRequest

		accessMock *accessMock
		mockArgs   *mockArgs

		wantErr bool
	}{
		{
			name: "ok",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", "user2"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
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
			name: "err_no_access",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "user",
				permissions: []string{},
			},
			wantErr: true,
		},
		{
			name: "err_invalid_role_id",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    0,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			wantErr: true,
		},
		{
			name: "err_empty_usernames",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			wantErr: true,
		},
		{
			name: "err_contains_empty_username",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1", ""},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.DeleteUsersFromRoleRequest{
				RoleID:    1,
				Usernames: []string{"user1"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionDeleteRoles},
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

			ctx := setupAccessMock(context.Background(), repo, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					DeleteUsersFromRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
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
