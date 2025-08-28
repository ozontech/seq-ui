package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdate(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "dashboard"
	dashboardMeta := "meta"

	type mockArgs struct {
		req types.UpdateDashboardRequest
		err error
	}

	tests := []struct {
		name string

		req      *dashboards.UpdateRequest
		want     *dashboards.UpdateResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Name: &dashboardName,
				Meta: &dashboardMeta,
			},
			want:     &dashboards.UpdateResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
			},
		},
		{
			name: "success_only_name",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Name: &dashboardName,
			},
			want:     &dashboards.UpdateResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
				},
			},
		},
		{
			name: "success_only_meta",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Meta: &dashboardMeta,
			},
			want:     &dashboards.UpdateResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Meta:      &dashboardMeta,
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name: "err_svc_invalid_uuid",
			req: &dashboards.UpdateRequest{
				Uuid: "invalid-uuid",
				Name: &dashboardName,
				Meta: &dashboardMeta,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc_empty_request",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_not_found",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Name: &dashboardName,
				Meta: &dashboardMeta,
			},
			wantCode: codes.NotFound,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
				err: types.ErrNotFound,
			},
		},
		{
			name: "err_repo_permission_denied",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Name: &dashboardName,
				Meta: &dashboardMeta,
			},
			wantCode: codes.PermissionDenied,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
				err: types.ErrPermissionDenied,
			},
		},
		{
			name: "err_repo_random",
			req: &dashboards.UpdateRequest{
				Uuid: dashboardUUID,
				Name: &dashboardName,
				Meta: &dashboardMeta,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Update(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.Update(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
