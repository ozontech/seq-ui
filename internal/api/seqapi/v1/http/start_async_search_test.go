package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeStartAsyncSearch(t *testing.T) {
	const (
		mockSearchID  = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockUserName  = "some_user"
		mockProfileID = 1
		meta          = `{"some":"meta"}`
	)

	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)

	formatReqBody := func(retention string) string {
		return fmt.Sprintf(`{"retention":%q,"query":%q,"from":%q,"to":%q,"with_docs":true,"size":100,"meta":"{\"some\":\"meta\"}","histogram":{"interval":"1s"},"aggregations":[{"field":"v","group_by":"level","agg_func":"avg","quantiles":[0.95],"interval":"30s"}]}`,
			retention, query, from.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	type mockArgs struct {
		proxyReq  *seqapi.StartAsyncSearchRequest
		proxyResp *seqapi.StartAsyncSearchResponse
		proxyErr  error

		profilesReq  types.GetOrCreateUserProfileRequest
		profilesResp types.UserProfile
		profilesErr  error

		repoReq types.SaveAsyncSearchRequest
		repoErr error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
	}{
		{
			name:    "ok",
			reqBody: formatReqBody("60s"),
			mockArgs: &mockArgs{
				proxyReq: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     query,
					From:      timestamppb.New(from),
					To:        timestamppb.New(to),
					WithDocs:  true,
					Size:      100,
					Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
						Interval: "1s",
					},
					Aggs: []*seqapi.AggregationQuery{
						{
							Field:     "v",
							GroupBy:   "level",
							Func:      seqapi.AggFunc_AGG_FUNC_AVG,
							Quantiles: []float64{0.95},
							Interval:  pointerTo("30s"),
						},
					},
					Meta: meta,
				},
				proxyResp: &seqapi.StartAsyncSearchResponse{
					SearchId: mockSearchID,
				},
				profilesReq: types.GetOrCreateUserProfileRequest{
					UserName: mockUserName,
				},
				profilesResp: types.UserProfile{
					ID:       mockProfileID,
					UserName: mockUserName,
				},
				repoReq: types.SaveAsyncSearchRequest{
					SearchID: mockSearchID,
					OwnerID:  mockProfileID,
					Meta:     meta,
				},
			},
			wantRespBody: `{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656"}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().StartAsyncSearch(gomock.Any(), tt.mockArgs.proxyReq).
					Return(tt.mockArgs.proxyResp, tt.mockArgs.proxyErr).Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				profilesRepoMock := mock_repo.NewMockUserProfiles(ctrl)
				profilesRepoMock.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.profilesReq).
					Return(tt.mockArgs.profilesResp, tt.mockArgs.profilesErr).Times(1)
				seqData.Mocks.ProfilesRepo = profilesRepoMock

				asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
				asyncSearchesRepoMock.EXPECT().SaveAsyncSearch(gomock.Any(), gomock.Any()).
					Return(tt.mockArgs.repoErr).Times(1)
				seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock
			}

			api := initTestAPIWithAsyncSearches(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/start", strings.NewReader(tt.reqBody))
			req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, mockUserName))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveStartAsyncSearch,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeStartAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)
	req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/start", strings.NewReader("{}"))

	httputil.DoTestHTTP(t, httputil.TestDataHTTP{
		Req:          req,
		Handler:      api.serveStartAsyncSearch,
		WantRespBody: `{"message":"async searches disabled"}`,
		WantStatus:   http.StatusBadRequest,
	})
}
