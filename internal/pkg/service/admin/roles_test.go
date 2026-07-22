package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

var (
	defaultSuperUser  = "superuser"
	adminCfg          = &config.Admin{SuperUsers: []string{defaultSuperUser}, CacheTTL: 5 * time.Minute}
	errSomethingWrong = errors.New("something happened wrong")
)

type accessMock struct {
	username    string
	permissions []string
	err         error
}

func setupAccessMock(ctx context.Context, repo *mock.MockAdmin, cache *mock_cache.MockCache, access *accessMock) context.Context {
	if access == nil {
		return ctx
	}

	ctx = types.SetUserKey(ctx, access.username)

	if access.username != defaultSuperUser {
		cache.EXPECT().
			Get(gomock.Any(), cacheKeyUserPerms+access.username).
			Return("", errors.New("not found")).
			Times(1)

		repo.EXPECT().
			GetUserPermissions(gomock.Any(), types.GetUserPermissionsRequest{Username: access.username}).
			Return(access.permissions, access.err).
			Times(1)

		cache.EXPECT().
			SetWithTTL(gomock.Any(), cacheKeyUserPerms+access.username, gomock.Any(), adminCfg.CacheTTL).
			Return(nil).
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

		wantRoleID int32
		wantErr    bool
	}{
		{
			name: "ok",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionRolesCreate},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesCreate},
			},
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionRolesCreate},
				},
				roleID: 1,
			},
			wantRoleID: 1,
		},
		{
			name: "ok_superuser",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionRolesCreate},
			},
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionRolesCreate},
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
				Permissions: []string{permissionRolesCreate},
			},
			wantErr: true,
		},
		{
			name: "err_no_access",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionRolesCreate},
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
				Permissions: []string{permissionRolesCreate},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesCreate},
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
				permissions: []string{permissionRolesCreate},
			},
			wantErr: true,
		},
		{
			name: "err_unknown_permissions",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{"unknown:operation"},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesCreate},
			},
			wantErr: true,
		},
		{
			name: "err_repo",
			req: types.CreateRoleRequest{
				Name:        "typical good boy",
				Permissions: []string{permissionRolesCreate},
			},
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesCreate},
			},
			mockArgs: &mockArgs{
				req: types.CreateRoleRequest{
					Name:        "typical good boy",
					Permissions: []string{permissionRolesCreate},
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
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					CreateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.roleID, tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					cache.EXPECT().
						Del(gomock.Any(), cacheKeyRoles).
						Times(1)
				}
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					AddUsersToRole(gomock.Any(), tt.req).
					Return(tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					for _, username := range tt.req.Usernames {
						cache.EXPECT().
							Del(gomock.Any(), cacheKeyUserPerms+username).
							Times(1)
					}
				}
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
				permissions: []string{permissionRolesRead},
			},
			mockArgs: &mockArgs{
				roles: []types.Role{
					{
						ID:          1,
						Name:        "admin",
						Permissions: []string{permissionRolesCreate, permissionRolesUpdate},
					},
				},
			},
			wantRoles: []types.Role{
				{
					ID:          1,
					Name:        "admin",
					Permissions: []string{permissionRolesCreate, permissionRolesUpdate},
				},
			},
		},
		{
			name: "ok_empty",
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesRead},
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
				permissions: []string{permissionRolesRead},
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
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				cache.EXPECT().
					Get(gomock.Any(), cacheKeyRoles).
					Return("", errors.New("not found")).
					Times(1)

				repo.EXPECT().
					GetRoles(gomock.Any()).
					Return(tt.mockArgs.roles, tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					cache.EXPECT().
						SetWithTTL(gomock.Any(), cacheKeyRoles, gomock.Any(), adminCfg.CacheTTL).
						Return(nil).
						Times(1)
				}
			}

			resp, err := svc.GetRoles(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRoles, resp.Roles)
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
				permissions: []string{permissionRolesRead},
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
				permissions: []string{permissionRolesRead},
			},
		},
		{
			name:    "err_repo",
			req:     types.GetRoleRequest{RoleID: 1},
			wantErr: true,
			accessMock: &accessMock{
				username:    "admin",
				permissions: []string{permissionRolesRead},
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
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

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
		req        types.UpdateRoleRequest
		err        error
		resetUsers []string
	}

	tests := []struct {
		name string
		req  types.UpdateRoleRequest

		mockArgs   *mockArgs
		accessMock *accessMock

		wantErr bool
	}{
		{
			name: "ok_name_and_permissions",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Name:        &name,
				Permissions: []string{permissionRolesRead},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Name:        &name,
					Permissions: []string{permissionRolesRead},
				},
				resetUsers: []string{"dima"},
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionRolesUpdate, permissionRolesRead},
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
				resetUsers: []string{"dima"},
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionRolesUpdate, permissionRolesRead},
			},
		},
		{
			name: "ok_permissions_only",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []string{permissionRolesRead},
			},
			mockArgs: &mockArgs{
				req: types.UpdateRoleRequest{
					RoleID:      1,
					Permissions: []string{permissionRolesRead},
				},
				resetUsers: []string{"dima"},
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionRolesUpdate, permissionRolesRead},
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
				resetUsers: []string{"dima"},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
			},
			wantErr: true,
		},
		{
			name: "err_unknown_permission",
			req: types.UpdateRoleRequest{
				RoleID:      1,
				Permissions: []string{"5:2"},
			},
			accessMock: &accessMock{
				username:    "typical good boy",
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					UpdateRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					cache.EXPECT().
						Del(gomock.Any(), cacheKeyRoles).
						Times(1)

					if len(tt.req.Permissions) > 0 {
						repo.EXPECT().
							GetRole(gomock.Any(), types.GetRoleRequest{RoleID: tt.req.RoleID}).
							Return(types.RoleInfo{Usernames: tt.mockArgs.resetUsers}, nil).
							Times(1)

						for _, u := range tt.mockArgs.resetUsers {
							cache.EXPECT().
								Del(gomock.Any(), cacheKeyUserPerms+u).
								Times(1)
						}
					}
				}
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
		req        types.DeleteRoleRequest
		err        error
		resetUsers []string
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
				permissions: []string{permissionRolesDelete, permissionRolesRead},
			},
			mockArgs: &mockArgs{
				req:        types.DeleteRoleRequest{RoleID: 1},
				resetUsers: []string{"dima"},
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
				permissions: []string{permissionRolesDelete, permissionRolesRead},
			},
			mockArgs: &mockArgs{
				req: types.DeleteRoleRequest{
					RoleID:            1,
					ReplacementRoleID: &replacementID,
				},
				resetUsers: []string{"dima"},
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
				permissions: []string{permissionRolesDelete, permissionRolesRead},
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
				permissions: []string{permissionRolesDelete, permissionRolesRead},
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
				permissions: []string{permissionRolesDelete, permissionRolesRead},
			},
			mockArgs: &mockArgs{
				req:        types.DeleteRoleRequest{RoleID: 1},
				err:        errSomethingWrong,
				resetUsers: []string{"dima"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetRole(gomock.Any(), types.GetRoleRequest{RoleID: tt.req.RoleID}).
					Return(types.RoleInfo{Usernames: tt.mockArgs.resetUsers}, nil).
					Times(1)

				for _, u := range tt.mockArgs.resetUsers {
					cache.EXPECT().
						Del(gomock.Any(), cacheKeyUserPerms+u).
						Times(1)
				}

				repo.EXPECT().
					DeleteRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					cache.EXPECT().
						Del(gomock.Any(), cacheKeyRoles).
						Times(1)
				}
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
				permissions: []string{permissionRolesUpdate},
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
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg)

			ctx := setupAccessMock(context.Background(), repo, cache, tt.accessMock)

			if tt.mockArgs != nil {
				repo.EXPECT().
					DeleteUsersFromRole(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					for _, username := range tt.mockArgs.req.Usernames {
						cache.EXPECT().
							Del(gomock.Any(), cacheKeyUserPerms+username).
							Times(1)
					}
				}
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
