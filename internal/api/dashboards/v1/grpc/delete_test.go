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

func TestDelete(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"

	type mockArgs struct {
		req types.DeleteDashboardRequest
		err error
	}

	tests := []struct {
		name string

		req      *dashboards.DeleteRequest
		want     *dashboards.DeleteResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &dashboards.DeleteRequest{
				Uuid: dashboardUUID,
			},
			want:     &dashboards.DeleteResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
					ProfileID: profileID,
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
			req: &dashboards.DeleteRequest{
				Uuid: "invalid-uuid",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_permission_denied",
			req: &dashboards.DeleteRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.PermissionDenied,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
					ProfileID: profileID,
				},
				err: types.ErrPermissionDenied,
			},
		},
		{
			name: "err_repo_random",
			req: &dashboards.DeleteRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
					ProfileID: profileID,
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
				mockedRepo.EXPECT().Delete(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.Delete(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
