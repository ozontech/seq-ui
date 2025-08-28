package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
)

func TestServeGetHist(t *testing.T) {
	var (
		service              = "service1"
		groupHash     uint64 = 123
		env                  = "prod"
		release              = "v1"
		twoMinutes           = 2 * time.Minute
		now                  = time.Now()
		oneMinuteAgo         = now.Add(-1 * time.Minute)
		twoMinutesAgo        = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req  types.GetErrorHistRequest
		resp []types.ErrorHistBucket
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
			name: "success_all_fields",
			reqBody: fmt.Sprintf(
				`{"service":"%s","group_hash":"%s","env":"%s","release":"%s","duration":"%s"}`,
				service, strconv.FormatUint(groupHash, 10), env, release, twoMinutes,
			),
			wantRespBody: fmt.Sprintf(
				`{"buckets":[{"time":"%s","count":10},{"time":"%s","count":20}]}`,
				oneMinuteAgo.Format(time.RFC3339Nano), twoMinutesAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					Service:   service,
					GroupHash: &groupHash,
					Env:       &env,
					Release:   &release,
					Duration:  &twoMinutes,
				},
				resp: []types.ErrorHistBucket{
					{Time: oneMinuteAgo, Count: 10},
					{Time: twoMinutesAgo, Count: 20},
				},
			},
		},
		{
			name:    "success_only_service_field",
			reqBody: fmt.Sprintf(`{"service":"%s"}`, service),
			wantRespBody: fmt.Sprintf(
				`{"buckets":[{"time":"%s","count":10},{"time":"%s","count":20}]}`,
				oneMinuteAgo.Format(time.RFC3339Nano), twoMinutesAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					Service: service,
				},
				resp: []types.ErrorHistBucket{
					{Time: oneMinuteAgo, Count: 10},
					{Time: twoMinutesAgo, Count: 20},
				},
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
				mockedRepo.EXPECT().GetErrorHist(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/errorgroups/v1/hist", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetHist,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
