package http

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
)

func TestServeGetServices(t *testing.T) {
	var (
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetServicesRequest

		services []string
		err      error
	}

	tests := []struct {
		name string

		req     getServicesRequest
		want    getServicesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: getServicesRequest{
				Query:  "test",
				Env:    &env,
				Limit:  10,
				Offset: 20,
			},
			want: getServicesResponse{
				Services: []string{"service1", "service2"},
			},

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query:  "test",
					Env:    &env,
					Limit:  10,
					Offset: 20,
				},

				services: []string{"service1", "service2"},
			},
		},
		{
			name: "err_svc",

			req:     getServicesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedSvc := svc_mock.NewMockService(ctrl)

			api := New(mockedSvc)

			if ma := tt.mockArgs; ma != nil {
				mockedSvc.EXPECT().
					GetServices(gomock.Any(), ma.req).
					Return(ma.services, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getServicesRequest, getServicesResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/services",
				Req:    tt.req,

				Handler: api.serveGetServices,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
