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

func TestSearch(t *testing.T) {
	var (
		userName = "unnamed"
	)

	type mockArgs struct {
		req  types.SearchDashboardsRequest
		resp types.DashboardInfosWithOwner
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.SearchRequest
		want     *dashboards.SearchResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &dashboards.SearchRequest{
				Query:  "test",
				Limit:  int32(testLimit),
				Offset: int32(testOffset),
			},
			want: &dashboards.SearchResponse{
				Dashboards: []*dashboards.SearchResponse_Dashboard{
					{Uuid: "064dc707-02b8-7000-8201-02a7f396738a", Name: "my test dashboard", OwnerName: "user1"},
					{Uuid: "064dc707-12b9-7000-a238-682b044c908b", Name: "tested", OwnerName: "user2"},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  "test",
					Limit:  testLimit,
					Offset: testOffset,
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "my test dashboard",
						},
						OwnerName: "user1",
					},
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-12b9-7000-a238-682b044c908b",
							Name: "tested",
						},
						OwnerName: "user2",
					},
				},
			},
		},
		{
			name: "ok_with_filter",
			req: &dashboards.SearchRequest{
				Query:  "test",
				Limit:  int32(testLimit),
				Offset: int32(testOffset),
				Filter: &dashboards.SearchRequest_Filter{
					OwnerName: &userName,
				},
			},
			want: &dashboards.SearchResponse{
				Dashboards: []*dashboards.SearchResponse_Dashboard{
					{Uuid: "064dc707-02b8-7000-8201-02a7f396738a", Name: "my test dashboard", OwnerName: userName},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  "test",
					Limit:  testLimit,
					Offset: testOffset,
					Filter: &types.SearchDashboardsFilter{
						OwnerName: &userName,
					},
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "my test dashboard",
						},
						OwnerName: userName,
					},
				},
			},
		},
		{
			name: "err_svc",
			req: &dashboards.SearchRequest{
				Query:  "test",
				Limit:  int32(testLimit),
				Offset: int32(testOffset),
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  "test",
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
					SearchDashboards(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.Search(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
