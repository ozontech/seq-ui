package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestServeGetGroups(t *testing.T) {
	var (
		service       = "service1"
		env           = "prod"
		release       = "v1"
		duration      = 2 * time.Minute
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req        types.GetErrorGroupsRequest
		groupsResp []types.ErrorGroup
		countsResp uint64
		err        error
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
				`{"service":"%s","env":"%s","release":"%s","duration":"%s","limit":10,"offset":0,"order":"%s","with_total":true}`,
				service, env, release, duration, OrderOldest,
			),
			wantRespBody: fmt.Sprintf(
				`{"total":2,"groups":[{"hash":"123","message":"some error 1","seen_total":10,"first_seen_at":"%s","last_seen_at":"%s","source":""},{"hash":"456","message":"some error 2","seen_total":5,"first_seen_at":"%s","last_seen_at":"%s","source":""}]}`,
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Env:       &env,
					Release:   &release,
					Duration:  &duration,
					Limit:     10,
					Offset:    0,
					Order:     types.OrderOldest,
					WithTotal: true,
				},
				groupsResp: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				countsResp: 2,
			},
		},
		{
			name:    "success_only_required_fields",
			reqBody: fmt.Sprintf(`{"service":"%s"}`, service),
			wantRespBody: fmt.Sprintf(
				`{"total":0,"groups":[{"hash":"123","message":"some error 1","seen_total":10,"first_seen_at":"%s","last_seen_at":"%s","source":""},{"hash":"456","message":"some error 2","seen_total":5,"first_seen_at":"%s","last_seen_at":"%s","source":""}]}`,
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
				twoMinutesAgo.Format(time.RFC3339Nano), oneMinuteAgo.Format(time.RFC3339Nano),
			),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Limit:   25,
					Offset:  0,
					Order:   types.OrderFrequent,
				},
				groupsResp: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
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
				mockedRepo.EXPECT().GetErrorGroups(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.groupsResp, tt.mockArgs.err).Times(1)
				if tt.mockArgs.req.WithTotal {
					mockedRepo.EXPECT().GetErrorGroupsCount(gomock.Any(), tt.mockArgs.req).
						Return(tt.mockArgs.countsResp, tt.mockArgs.err).Times(1)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/errorgroups/v1/groups", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetGroups,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
