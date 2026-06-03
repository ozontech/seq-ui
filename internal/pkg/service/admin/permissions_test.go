package admin

import (
	"context"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCheckAccess(t *testing.T) {
	type mockArgs struct {
		req   types.GetUserPermissionsRequest
		perms uint64
		err   error
	}

	tests := []struct {
		name         string
		username     string
		requiredPerm uint64
		wantErr      error
		mockArgs     *mockArgs
	}{
		{
			name:         "ok_super_user",
			username:     defaultSuperUser,
			requiredPerm: permissionManageRoles,
		},
		{
			name:         "err_no_auth",
			requiredPerm: permissionManageRoles,
			wantErr:      types.ErrUnauthenticated,
		},
		{
			name:         "err_permission_denied",
			username:     "typical bad boy",
			requiredPerm: permissionManageRoles,
			wantErr:      types.ErrPermissionDenied,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "typical bad boy"},
				perms: 0,
			},
		},
		{
			name:         "ok_allowed",
			username:     "typical good boy",
			requiredPerm: permissionManageRoles,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "typical good boy"},
				perms: permissionManageRoles,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			repo := mock.NewMockAdmin(ctrl)
			svc := New(repo, adminCfg).(*service)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetUserPermissions(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.perms, tt.mockArgs.err).
					Times(1)
			}

			ctx := t.Context()
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
		perms uint64
		err   error
	}

	tests := []struct {
		name      string
		req       types.GetUserPermissionsRequest
		wantPerms uint64
		wantErr   bool
		mockArgs  *mockArgs
	}{
		{
			name:      "ok",
			req:       types.GetUserPermissionsRequest{Username: "user1"},
			wantPerms: permissionManageRoles,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "user1"},
				perms: permissionManageRoles,
			},
		},
		{
			name:      "ok_no_permissions",
			req:       types.GetUserPermissionsRequest{Username: "user1"},
			wantPerms: 0,
			mockArgs: &mockArgs{
				req:   types.GetUserPermissionsRequest{Username: "user1"},
				perms: 0,
			},
		},
		{
			name:      "err_repo",
			req:       types.GetUserPermissionsRequest{Username: "user1"},
			wantPerms: 0,
			wantErr:   true,
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
			svc := New(repo, adminCfg)

			if tt.mockArgs != nil {
				repo.EXPECT().
					GetUserPermissions(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.perms, tt.mockArgs.err).
					Times(1)
			}

			// first call goes to repo.
			permsFromRepo, err := svc.GetUserPermissions(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPerms, permsFromRepo)

			// second call served from cache.
			permsFromCache, err := svc.GetUserPermissions(context.Background(), tt.req)
			require.NoError(t, err)
			require.Equal(t, permsFromRepo, permsFromCache)
		})
	}
}

func TestGetAvailablePermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockAdmin(ctrl)
	svc := New(repo, adminCfg)

	got := svc.GetAvailablePermissions()
	require.Equal(t, availablePermissions, got)
}

func TestMaskUnmaskPermissions(t *testing.T) {
	tests := []struct {
		name  string
		perms []uint64
		mask  uint64
	}{
		{
			name:  "single_permission",
			perms: []uint64{permissionManageRoles},
			mask:  permissionManageRoles,
		},
		{
			name:  "empty_permission",
			perms: []uint64{},
			mask:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			masked := maskPermissions(tt.perms)
			require.Equal(t, tt.mask, masked)

			unmasked := unmaskPermissions(masked)
			require.Equal(t, tt.perms, unmasked)
		})
	}
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name    string
		perms   []uint64
		wantErr bool
	}{
		{
			name:  "ok",
			perms: []uint64{permissionManageRoles},
		},
		{
			name:    "err_empty",
			perms:   []uint64{},
			wantErr: true,
		},
		{
			name:    "err_unknown",
			perms:   []uint64{52},
			wantErr: true,
		},
		{
			name:    "err_mixed",
			perms:   []uint64{permissionManageRoles, 52},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePermissions(tt.perms)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}
