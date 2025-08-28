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

func TestServeGetReleases(t *testing.T) {
	var (
		service = "service1"
		env     = "prod"
	)

	type mockArgs struct {
		req  types.GetErrorGroupReleasesRequest
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
			reqBody:      `{"service":"service1","env":"prod"}`,
			wantRespBody: `{"releases":["v1","v2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
					Env:     &env,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name:         "success_no_env_field",
			reqBody:      `{"service":"service1","group_hash":"123"}`,
			wantRespBody: `{"releases":["v1","v2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name:         "success_only_service_field",
			reqBody:      `{"service":"service1"}`,
			wantRespBody: `{"releases":["v1","v2"]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name:         "err_no_service_field",
			reqBody:      `{"group_hash":"123"}`,
			wantRespBody: `{"message":"invalid request field: 'service' must not be empty"}`,
			wantStatus:   http.StatusBadRequest,
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
				mockedRepo.EXPECT().GetErrorReleases(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/errorgroups/v1/releases", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetReleases,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
