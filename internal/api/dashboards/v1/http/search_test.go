package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeSearch(t *testing.T) {
	userName := "unnamed"
	query := "test-query"

	type mockArgs struct {
		req  types.SearchDashboardsRequest
		resp types.DashboardInfosWithOwner
		err  error
	}

	tests := []struct {
		name string

		req     searchRequest
		want    searchResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  searchRequest{Query: query, Limit: testLimit, Offset: testOffset},
			want: searchResponse{
				Dashboards: infosWithOwner{
					{info: info{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "my test dashboard"}, OwnerName: "user1"},
					{info: info{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "tested"}, OwnerName: "user2"},
				},
			},
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
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
			name: "ok_filter",
			req: searchRequest{
				Query:  query,
				Limit:  testLimit,
				Offset: testOffset,
				Filter: &searchFilter{OwnerName: &userName},
			},
			want: searchResponse{
				Dashboards: infosWithOwner{
					{info: info{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "my test dashboard"}, OwnerName: userName},
				},
			},
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
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
			name:    "err_svc",
			req:     searchRequest{Query: query, Limit: testLimit, Offset: testOffset},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[searchRequest, searchResponse]{
				Method:  http.MethodPost,
				Target:  "/dashboards/v1/search",
				Req:     tt.req,
				Handler: api.serveSearch,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
