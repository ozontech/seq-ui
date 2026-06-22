package http

import (
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetByUUID(t *testing.T) {
	type mockArgs struct {
		uuid string
		resp types.Dashboard
		err  error
	}

	tests := []struct {
		name string

		want    dashboard
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: dashboard{
				Name:      "dashboard1",
				Meta:      "meta1",
				OwnerName: "owner",
			},
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				resp: types.Dashboard{
					OwnerName: "owner",
					Name:      "dashboard1",
					Meta:      "meta1",
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetDashboardByUUID(gomock.Any(), tt.mockArgs.uuid).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, dashboard]{
				Method:  http.MethodGet,
				Target:  fmt.Sprintf("/dashboards/v1/%s", dashboardUUID),
				Handler: withUUID(api.serveGetByUUID, dashboardUUID),
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
