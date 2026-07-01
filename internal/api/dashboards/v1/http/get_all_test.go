package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetAll(t *testing.T) {
	type mockArgs struct {
		req  types.GetAllDashboardsRequest
		resp types.DashboardInfosWithOwner
		err  error
	}

	tests := []struct {
		name string

		req     getAllRequest
		want    getAllResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  getAllRequest{Limit: testLimit, Offset: testOffset},
			want: getAllResponse{
				Dashboards: infosWithOwner{
					{info: info{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"}, OwnerName: "user1"},
					{info: info{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"}, OwnerName: "user2"},
				},
			},
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
			name:    "err_svc",
			req:     getAllRequest{Limit: testLimit, Offset: testOffset},
			wantErr: true,
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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getAllRequest, getAllResponse]{
				Method:  http.MethodPost,
				Target:  "/dashboards/v1/all",
				Req:     tt.req,
				Handler: api.serveGetAll,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
