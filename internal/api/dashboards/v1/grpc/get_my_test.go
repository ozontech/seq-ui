package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
)

func TestGetMy(t *testing.T) {
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
	}{
		{
			name: "ok",
			req: &dashboards.GetMyRequest{
				Limit:  int32(limit),
				Offset: int32(offset),
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
					Limit:  limit,
					Offset: offset,
				},
				resp: types.DashboardInfos{
					{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"},
					{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"},
				},
			},
		},
		{
			name: "err_svc",
			req: &dashboards.GetMyRequest{
				Limit:  int32(limit),
				Offset: int32(offset),
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetUserDashboardsRequest{
					Limit:  limit,
					Offset: offset,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetMyDashboards(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.GetMy(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
