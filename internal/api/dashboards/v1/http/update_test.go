package http

import (
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeUpdate(t *testing.T) {
	type mockArgs struct {
		req types.UpdateDashboardRequest
		err error
	}

	tests := []struct {
		name string

		req     updateRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  updateRequest{Name: &dashboardName, Meta: &dashboardMeta},
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					UUID: dashboardUUID,
					Name: &dashboardName,
					Meta: &dashboardMeta,
				},
			},
		},
		{
			name:    "err_svc",
			req:     updateRequest{Name: &dashboardName, Meta: &dashboardMeta},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					UUID: dashboardUUID,
					Name: &dashboardName,
					Meta: &dashboardMeta,
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
				mockedSvc.EXPECT().UpdateDashboard(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[updateRequest, struct{}]{
				Method:  http.MethodPatch,
				Target:  fmt.Sprintf("/dashboards/v1/%s", dashboardUUID),
				Req:     tt.req,
				Handler: withUUID(api.serveUpdate, dashboardUUID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}
