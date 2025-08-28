package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
)

func TestServeGetServices(t *testing.T) {
	var (
		query = "itc"
		env   = "prod"
	)

	type mockArgs struct {
		req  types.GetServicesRequest
		resp []string
		err  error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
	}{
		{
			name:         "success_all_fields",
			reqBody:      `{"query":"itc","env":"prod"}`,
			wantRespBody: `{"services":["service1","service2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query: query,
					Env:   &env,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name:         "success_no_env_field",
			reqBody:      `{"query":"itc"}`,
			wantRespBody: `{"services":["service1","service2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query: query,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name:         "success_no_query_field",
			reqBody:      `{"query":"","env":"prod"}`,
			wantRespBody: `{"services":["service1","service2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Env: &env,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name:         "success_no_query_no_env",
			reqBody:      `{"query":""}`,
			wantRespBody: `{"services":["service1","service2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req:  types.GetServicesRequest{},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := repo_mock.NewMockRepository(ctrl)
			api := New(errorgroups.New(mockedRepo, config.LogTagsMapping{}))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetServices(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/errorgroups/v1/services", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetServices,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
