package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
)

func TestCheckAccess(t *testing.T) {
	type mockArgs struct {
		req   types.GetUserPermissionsRequest
		perms []string
		err   error
	}

	tests := []struct {
		name string

		username     string
		requiredPerm string
		wantErr      error

		mockArgs *mockArgs
	}{
		{
			name:         "ok_super_user",
			username:     defaultSuperUser,
			requiredPerm: permissionRolesCreate,
		},
		{
			name:         "err_no_auth",
			requiredPerm: permissionRolesCreate,
			wantErr:      types.ErrUnauthenticated,
		},
		{
			name:         "err_permission_denied",
			username:     "typical bad boy",
			requiredPerm: permissionRolesCreate,
			wantErr:      types.ErrPermissionDenied,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "typical bad boy"},
				perms: []string{},
			},
		},
		{
			name:         "ok_allowed",
			username:     "typical good boy",
			requiredPerm: permissionRolesCreate,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "typical good boy"},
				perms: []string{permissionRolesCreate},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg).(*service)

			if tt.mockArgs != nil {
				cache.EXPECT().
					Get(gomock.Any(), cacheKeyUserPerms+tt.mockArgs.req.Username).
					Return("", errors.New("not found")).
					Times(1)
				repo.EXPECT().
					GetUserPermissions(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.perms, tt.mockArgs.err).
					Times(1)
				cache.EXPECT().
					SetWithTTL(gomock.Any(), cacheKeyUserPerms+tt.mockArgs.req.Username, gomock.Any(), adminCfg.CacheTTL).
					Return(nil).
					Times(1)
			}

			ctx := context.Background()
			if tt.username != "" {
				ctx = types.SetUserKey(ctx, tt.username)
			}

			err := svc.checkAccess(ctx, tt.requiredPerm)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetUserPermissions(t *testing.T) {
	type mockArgs struct {
		req   types.GetUserPermissionsRequest
		perms []string
		err   error
	}

	tests := []struct {
		name      string
		req       types.GetUserPermissionsRequest
		wantPerms []string
		wantErr   bool
		mockArgs  *mockArgs
	}{
		{
			name:      "ok",
			req:       types.GetUserPermissionsRequest{Username: "user1"},
			wantPerms: []string{permissionRolesCreate, permissionRolesDelete},
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "user1"},
				perms: []string{permissionRolesCreate, permissionRolesDelete},
			},
		},
		{
			name:      "ok_no_permissions",
			req:       types.GetUserPermissionsRequest{Username: "user1"},
			wantPerms: []string{},
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "user1"},
				perms: []string{},
			},
		},
		{
			name:    "err_repo",
			req:     types.GetUserPermissionsRequest{Username: "user1"},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.GetUserPermissionsRequest{Username: "user1"},
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
			svc := New(repo, cache, adminCfg).(*service)

			cache.EXPECT().
				Get(gomock.Any(), cacheKeyUserPerms+tt.mockArgs.req.Username).
				Return("", errors.New("not found")).
				Times(1)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetUserPermissions(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.perms, tt.mockArgs.err).
					Times(1)

				if tt.mockArgs.err == nil {
					cache.EXPECT().
						SetWithTTL(gomock.Any(), cacheKeyUserPerms+tt.mockArgs.req.Username, gomock.Any(), adminCfg.CacheTTL).
						Return(nil).
						Times(1)
				}
			}

			perms, err := svc.getUserPermissions(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPerms, perms)
		})
	}
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name    string
		perms   []string
		wantErr bool
	}{
		{
			name:  "ok",
			perms: []string{permissionRolesCreate},
		},
		{
			name:    "err_empty",
			perms:   []string{},
			wantErr: true,
		},
		{
			name:    "err_unknown",
			perms:   []string{"roles:unknownOperation"},
			wantErr: true,
		},
		{
			name:    "err_mixed",
			perms:   []string{permissionRolesCreate, "roles:unknownOperation"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			cache := mock_cache.NewMockCache(ctrl)
			svc := New(repo, cache, adminCfg).(*service)

			err := svc.validatePermissions(tt.perms)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
