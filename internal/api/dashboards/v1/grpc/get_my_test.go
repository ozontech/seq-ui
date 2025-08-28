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

func TestGetMy(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1

	type mockArgs struct {
		req  types.GetUserDashboardsRequest
		resp types.DashboardInfos
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.GetMyRequest
		want     *dashboards.GetMyResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &dashboards.GetMyRequest{
				Limit:  2,
				Offset: 0,
			},
			want: &dashboards.GetMyResponse{
				Dashboards: []*dashboards.GetMyResponse_Dashboard{
					{Uuid: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"},
					{Uuid: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetUserDashboardsRequest{
					ProfileID: profileID,
					Limit:     2,
					Offset:    0,
				},
				resp: types.DashboardInfos{
					{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"},
					{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"},
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name: "err_svc_invalid_limit",
			req: &dashboards.GetMyRequest{
				Limit:  0,
				Offset: 0,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc_invalid_offset",
			req: &dashboards.GetMyRequest{
				Limit:  2,
				Offset: -10,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_random",
			req: &dashboards.GetMyRequest{
				Limit:  2,
				Offset: 0,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetUserDashboardsRequest{
					ProfileID: profileID,
					Limit:     2,
					Offset:    0,
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
				mockedRepo.EXPECT().GetMy(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.GetMy(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
