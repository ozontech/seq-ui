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

func TestServeGetDetails(t *testing.T) {
	var (
		service              = "service1"
		groupHash     uint64 = 123
		env                  = "prod"
		release              = "test_release"
		now                  = time.Now()
		oneMinuteAgo         = now.Add(-1 * time.Minute)
		twoMinutesAgo        = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req         types.GetErrorGroupDetailsRequest
		detailsResp types.ErrorGroupDetails
		countsResp  *types.ErrorGroupCounts
		err         error
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
				`{"service":"%s","group_hash":"%d","env":"%s","release":"%s"}`,
				service, groupHash, env, release,
			),
			wantRespBody: fmt.Sprintf(
				`{"group_hash":"123","message":"some error","seen_total":10,"first_seen_at":"%s","last_seen_at":"%s","source":"","distributions":{"by_env":[{"value":"prod","percent":100}],"by_release":[{"value":"test_release","percent":100}]},"envs":[{"env":"prod","percent":100}]}`,
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					Service:   service,
					GroupHash: groupHash,
					Env:       &env,
					Release:   &release,
				},
				detailsResp: types.ErrorGroupDetails{
					GroupHash:   123,
					Message:     "some error",
					SeenTotal:   10,
					FirstSeenAt: twoMinutesAgo,
					LastSeenAt:  oneMinuteAgo,
				},
			},
		},
		{
			name: "success_only_required_fields",
			reqBody: fmt.Sprintf(
				`{"service":"%s","group_hash":"%s"}`,
				service, strconv.FormatUint(groupHash, 10),
			),
			wantRespBody: fmt.Sprintf(
				`{"group_hash":"123","message":"some error","seen_total":10,"first_seen_at":"%s","last_seen_at":"%s","source":"","distributions":{"by_env":[{"value":"env1","percent":70},{"value":"env2","percent":30}],"by_release":[{"value":"release2","percent":60},{"value":"release1","percent":40}]},"envs":[{"env":"env1","percent":70},{"env":"env2","percent":30}]}`,
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					Service:   service,
					GroupHash: groupHash,
				},
				detailsResp: types.ErrorGroupDetails{
					GroupHash:   123,
					Message:     "some error",
					SeenTotal:   10,
					FirstSeenAt: twoMinutesAgo,
					LastSeenAt:  oneMinuteAgo,
				},
				countsResp: &types.ErrorGroupCounts{
					ByEnv: types.ErrorGroupCount{
						"env1": 7,
						"env2": 3,
					},
					ByRelease: types.ErrorGroupCount{
						"release1": 4,
						"release2": 6,
					},
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
			name:         "err_no_group_hash_field",
			reqBody:      `{"service":"service1"}`,
			wantRespBody: `{"message":"failed to parse group_hash: strconv.ParseUint: parsing \"\": invalid syntax"}`,
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
				mockedRepo.EXPECT().GetErrorDetails(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.detailsResp, tt.mockArgs.err).Times(1)
				if tt.mockArgs.countsResp != nil {
					mockedRepo.EXPECT().GetErrorCounts(gomock.Any(), tt.mockArgs.req).
						Return(*tt.mockArgs.countsResp, tt.mockArgs.err).Times(1)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/errorgroups/v1/details", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetDetails,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
