package http

import (
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

func TestServeFetchAsyncSearchResult(t *testing.T) {
	var (
		mockSearchID = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockTime     = time.Date(2025, 8, 6, 17, 52, 12, 123, time.UTC)
		meta         = `{"some":"meta"}`
	)

	type mockArgs struct {
		proxyReq  *seqapi.FetchAsyncSearchResultRequest
		proxyResp *seqapi.FetchAsyncSearchResultResponse
		proxyErr  error

		repoResp types.AsyncSearchInfo
		repoErr  error
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
			reqBody: `{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656","limit":2,"offset":10,"order":"desc"}`,
			mockArgs: &mockArgs{
				proxyReq: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: mockSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
				proxyResp: &seqapi.FetchAsyncSearchResultResponse{
					Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
					Request: &seqapi.StartAsyncSearchRequest{
						Retention: durationpb.New(60 * time.Second),
						Query:     "message:error",
						From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
						To:        timestamppb.New(mockTime),
						Aggs: []*seqapi.AggregationQuery{
							{
								Field:     "x",
								GroupBy:   "level",
								Func:      seqapi.AggFunc_AGG_FUNC_AVG,
								Quantiles: []float64{0.9, 0.5},
							},
						},
						Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
							Interval: "1s",
						},
						WithDocs: true,
						Size:     100,
					},
					Response: &seqapi.SearchResponse{
						Events: []*seqapi.Event{
							{
								Id: "017a854298010000-850287cfa326a7fc",
								Data: map[string]string{
									"level":   "3",
									"message": "some error",
									"x":       "2",
								},
								Time: timestamppb.New(mockTime.Add(-1 * time.Minute)),
							},
							{
								Id: "017a854298010000-8502fe7f2aa33df3",
								Data: map[string]string{
									"level":   "2",
									"message": "some error 2",
									"x":       "8",
								},
								Time: timestamppb.New(mockTime.Add(-2 * time.Minute)),
							},
						},
						Total: 2,
						Histogram: &seqapi.Histogram{
							Buckets: []*seqapi.Histogram_Bucket{
								{
									DocCount: 7,
									Key:      1,
								},
								{
									DocCount: 9,
									Key:      2,
								},
							},
						},
						Aggregations: []*seqapi.Aggregation{
							{
								Buckets: []*seqapi.Aggregation_Bucket{
									{
										Key:       "3",
										Value:     pointerTo(2),
										NotExists: 0,
										Quantiles: []float64{2, 1},
										Ts:        timestamppb.New(mockTime),
									},
									{
										Key:       "2",
										Value:     pointerTo(8),
										NotExists: 1,
										Quantiles: []float64{7, 4},
										Ts:        timestamppb.New(mockTime.Add(-1 * time.Minute)),
									},
								},
								NotExists: 2,
							},
						},
						Error: &seqapi.Error{
							Code:    seqapi.ErrorCode_ERROR_CODE_NO,
							Message: "some error",
						},
					},
					StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
					ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
					Progress:  1,
					DiskUsage: 512,
				},
				repoResp: types.AsyncSearchInfo{
					SearchID: mockSearchID,
					Meta:     meta,
				},
			},
			wantRespBody: `{"status":"done","request":{"retention":"seconds:60","query":"message:error","from":"2025-08-06T17:37:12.000000123Z","to":"2025-08-06T17:52:12.000000123Z","aggregations":[{"field":"x","group_by":"level","agg_func":"avg","quantiles":[0.9,0.5]}],"histogram":{"interval":"1s"},"with_docs":true,"size":100},"response":{"events":[{"id":"017a854298010000-850287cfa326a7fc","data":{"level":"3","message":"some error","x":"2"},"time":"2025-08-06T17:51:12.000000123Z"},{"id":"017a854298010000-8502fe7f2aa33df3","data":{"level":"2","message":"some error 2","x":"8"},"time":"2025-08-06T17:50:12.000000123Z"}],"histogram":{"buckets":[{"key":"1","docCount":"7"},{"key":"2","docCount":"9"}]},"aggregations":[{"buckets":[{"key":"3","value":2,"quantiles":[2,1]},{"key":"2","value":8,"not_exists":1,"quantiles":[7,4]}],"not_exists":2}],"total":"2","error":{"code":"ERROR_CODE_NO","message":"some error"},"partialResponse":false},"started_at":"2025-08-06T17:51:42.000000123Z","expires_at":"2025-08-06T17:52:42.000000123Z","progress":1,"disk_usage":"512","meta":"{\"some\":\"meta\"}"}`,
			wantStatus:   http.StatusOK,
		},
		{
			name:         "err_limit",
			reqBody:      `{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656","limit":-10,"offset":20}`,
			wantRespBody: `{"message":"invalid request field: 'limit' must be non-negative"}`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "err_offset",
			reqBody:      `{"search_id":"c9a34cf8-4c66-484e-9cc2-42979d848656","limit":10,"offset":-20}`,
			wantRespBody: `{"message":"invalid request field: 'offset' must be non-negative"}`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "invalid id",
			reqBody:      `{"search_id":"some_invalid_id"}`,
			wantRespBody: `{"message":"invalid request field: invalid uuid"}`,
			wantStatus:   http.StatusBadRequest,
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

				asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
				asyncSearchesRepoMock.EXPECT().GetAsyncSearchById(gomock.Any(), mockSearchID).
					Return(tt.mockArgs.repoResp, tt.mockArgs.repoErr).Times(1)
				seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().FetchAsyncSearchResult(gomock.Any(), tt.mockArgs.proxyReq).
					Return(tt.mockArgs.proxyResp, tt.mockArgs.proxyErr).Times(1)
				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPIWithAsyncSearches(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/fetch", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveFetchAsyncSearchResult,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeFetchAsyncSearchResult_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)
	req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/async_search/fetch", strings.NewReader("{}"))

	httputil.DoTestHTTP(t, httputil.TestDataHTTP{
		Req:          req,
		Handler:      api.serveFetchAsyncSearchResult,
		WantRespBody: `{"message":"async searches disabled"}`,
		WantStatus:   http.StatusBadRequest,
	})
}

func pointerTo[T float64](in T) *T {
	return &in
}
