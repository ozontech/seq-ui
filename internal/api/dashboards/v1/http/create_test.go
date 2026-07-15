package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeCreate(t *testing.T) {
	var (
		dashboardUUID = "064dc707-02b8-7000-8201-02a7f396738a"
		dashboardMeta = "my_meta"
		dashboardName = "my_dashboard"
	)

	type mockArgs struct {
		req  types.CreateDashboardRequest
		resp string
		err  error
	}

	tests := []struct {
		name string

		req     createRequest
		want    createResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  createRequest{Name: dashboardName, Meta: dashboardMeta},
			want: createResponse{UUID: dashboardUUID},
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					Name: dashboardName,
					Meta: dashboardMeta,
				},
				resp: dashboardUUID,
			},
		},
		{
			name:    "err_svc",
			req:     createRequest{Name: dashboardName, Meta: dashboardMeta},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					Name: dashboardName,
					Meta: dashboardMeta,
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
					CreateDashboard(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[createRequest, createResponse]{
				Method:  http.MethodPost,
				Target:  "/dashboards/v1/",
				Req:     tt.req,
				Handler: api.serveCreate,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
