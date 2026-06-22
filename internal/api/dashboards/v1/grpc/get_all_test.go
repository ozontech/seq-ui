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

func TestGetAll(t *testing.T) {
	type mockArgs struct {
		req  types.GetAllDashboardsRequest
		resp types.DashboardInfosWithOwner
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.GetAllRequest
		want     *dashboards.GetAllResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &dashboards.GetAllRequest{
				Limit:  int32(testLimit),
				Offset: int32(testOffset),
			},
			want: &dashboards.GetAllResponse{
				Dashboards: []*dashboards.GetAllResponse_Dashboard{
					{Uuid: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1", OwnerName: "user1"},
					{Uuid: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2", OwnerName: "user2"},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetAllDashboardsRequest{
					Limit:  testLimit,
					Offset: testOffset,
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "dashboard1",
						},
						OwnerName: "user1",
					},
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-12b9-7000-a238-682b044c908b",
							Name: "dashboard2",
						},
						OwnerName: "user2",
					},
				},
			},
		},
		{
			name: "err_svc",
			req: &dashboards.GetAllRequest{
				Limit:  int32(testLimit),
				Offset: int32(testOffset),
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetAllDashboardsRequest{
					Limit:  testLimit,
					Offset: testOffset,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetAllDashboards(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.GetAll(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
