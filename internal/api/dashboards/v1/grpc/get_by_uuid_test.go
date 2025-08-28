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

func TestGetByUUID(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "my_dashboard"
	dashboardMeta := "my_meta"
	dashboardOwner := "owner"

	type mockArgs struct {
		uuid string
		resp types.Dashboard
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.GetByUUIDRequest
		want     *dashboards.GetByUUIDResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &dashboards.GetByUUIDRequest{
				Uuid: dashboardUUID,
			},
			want: &dashboards.GetByUUIDResponse{
				Name:      dashboardName,
				Meta:      dashboardMeta,
				OwnerName: dashboardOwner,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				resp: types.Dashboard{
					Name:      dashboardName,
					Meta:      dashboardMeta,
					OwnerName: dashboardOwner,
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
			req: &dashboards.GetByUUIDRequest{
				Uuid: "invalid-uuid",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_not_found",
			req: &dashboards.GetByUUIDRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.NotFound,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  types.ErrNotFound,
			},
		},
		{
			name: "err_repo_random",
			req: &dashboards.GetByUUIDRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetByUUID(gomock.Any(), tt.mockArgs.uuid).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.GetByUUID(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
