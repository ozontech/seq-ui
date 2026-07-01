package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetMy(t *testing.T) {
	type mockArgs struct {
		req  types.GetUserDashboardsRequest
		resp types.DashboardInfos
		err  error
	}

	tests := []struct {
		name string

		req     getMyRequest
		want    getMyResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  getMyRequest{Limit: testLimit, Offset: testOffset},
			want: getMyResponse{
				Dashboards: infos{
					{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"},
					{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"},
				},
			},
			mockArgs: &mockArgs{
				req: types.GetUserDashboardsRequest{
					Limit:  testLimit,
					Offset: testOffset,
				},
				resp: types.DashboardInfos{
					{UUID: "064dc707-02b8-7000-8201-02a7f396738a", Name: "dashboard1"},
					{UUID: "064dc707-12b9-7000-a238-682b044c908b", Name: "dashboard2"},
				},
			},
		},
		{
			name:    "err_svc",
			req:     getMyRequest{Limit: testLimit, Offset: testOffset},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.GetUserDashboardsRequest{
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

			api, mockedRepo := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().
					GetMyDashboards(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getMyRequest, getMyResponse]{
				Method:  http.MethodPost,
				Target:  "/dashboards/v1/my",
				Req:     tt.req,
				Handler: api.serveGetMy,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
