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

func TestServeGetServiceFilters(t *testing.T) {
	var (
		service = "test-service"
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetServiceFiltersRequest

		filters []types.ServiceFilter
		err     error
	}

	tests := []struct {
		name string

		req     getServiceFiltersRequest
		want    getServiceFiltersResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: getServiceFiltersRequest{
				Service: service,
				Env:     &env,
			},
			want: getServiceFiltersResponse{
				Filters: []serviceFilter{
					{Key: "filter1", Values: []string{"val1", "val2"}},
					{Key: "filter2", Values: []string{"val10", "val20"}},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetServiceFiltersRequest{
					Service: service,
					Env:     &env,
				},

				filters: []types.ServiceFilter{
					{Key: "filter1", Values: []string{"val1", "val2"}},
					{Key: "filter2", Values: []string{"val10", "val20"}},
				},
			},
		},
		{
			name: "err_svc",

			req:     getServiceFiltersRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetServiceFiltersRequest{},

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
					GetServiceFilters(gomock.Any(), ma.req).
					Return(ma.filters, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getServiceFiltersRequest, getServiceFiltersResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/service_filters",
				Req:    tt.req,

				Handler: api.serveGetServiceFilters,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
