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

func TestCreate(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "my_dashboard"
	dashboardMeta := "my_meta"

	type mockArgs struct {
		req  types.CreateDashboardRequest
		resp string
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.CreateRequest
		want     *dashboards.CreateResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &dashboards.CreateRequest{
				Name: dashboardName,
				Meta: dashboardMeta,
			},
			want: &dashboards.CreateResponse{
				Uuid: dashboardUUID,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					ProfileID: profileID,
					Name:      dashboardName,
					Meta:      dashboardMeta,
				},
				resp: dashboardUUID,
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name: "err_svc_empty_name",
			req: &dashboards.CreateRequest{
				Meta: dashboardMeta,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc_empty_meta",
			req: &dashboards.CreateRequest{
				Name: dashboardName,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_random",
			req: &dashboards.CreateRequest{
				Name: dashboardName,
				Meta: dashboardMeta,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					ProfileID: profileID,
					Name:      dashboardName,
					Meta:      dashboardMeta,
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
				mockedRepo.EXPECT().Create(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.Create(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
