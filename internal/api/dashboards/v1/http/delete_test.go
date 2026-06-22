package http

import (
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeDelete(t *testing.T) {
	type mockArgs struct {
		req types.DeleteDashboardRequest
		err error
	}

	tests := []struct {
		name string

		wantErr  bool
		mockArgs *mockArgs
	}{
		{
			name: "ok",
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID: dashboardUUID,
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID: dashboardUUID,
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
					DeleteDashboard(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
				Method:  http.MethodDelete,
				Target:  fmt.Sprintf("/dashboards/v1/%s", dashboardUUID),
				Handler: withUUID(api.serveDelete, dashboardUUID),
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}
