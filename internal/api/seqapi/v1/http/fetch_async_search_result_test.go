package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeFetchAsyncSearchResult(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.FetchAsyncSearchResultRequest
		resp *seqapi.FetchAsyncSearchResultResponse
		err  error
	}

	tests := []struct {
		name string

		req     fetchAsyncSearchResultRequest
		want    fetchAsyncSearchResultResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: fetchAsyncSearchResultRequest{
				SearchID: testSearchID,
				Limit:    2,
				Offset:   10,
				Order:    oDESC,
			},
			want: fetchAsyncSearchResultResponseFromProto(&seqapi.FetchAsyncSearchResultResponse{
				Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
				Request: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     "message:error",
					From:      timestamppb.New(testSomeMoment.Add(-15 * time.Minute)),
					To:        timestamppb.New(testSomeMoment),
					Aggs: []*seqapi.AggregationQuery{
						{
							Field:     "x",
							GroupBy:   "level",
							Func:      seqapi.AggFunc_AGG_FUNC_AVG,
							Quantiles: []float64{0.9, 0.5},
						},
						{
							Field:    "y",
							GroupBy:  "level",
							Func:     seqapi.AggFunc_AGG_FUNC_SUM,
							Interval: pointerTo("30s"),
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
							Time: timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
						},
						{
							Id: "017a854298010000-8502fe7f2aa33df3",
							Data: map[string]string{
								"level":   "2",
								"message": "some error 2",
								"x":       "8",
							},
							Time: timestamppb.New(testSomeMoment.Add(-2 * time.Minute)),
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
									Value:     pointerTo[float64](2),
									NotExists: 0,
									Quantiles: []float64{2, 1},
								},
								{
									Key:       "2",
									Value:     pointerTo[float64](8),
									NotExists: 1,
									Quantiles: []float64{7, 4},
								},
							},
						},
						{
							Buckets: []*seqapi.Aggregation_Bucket{
								{
									Key:       "33",
									Value:     pointerTo[float64](2),
									NotExists: 0,
									Ts:        timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
								},
								{
									Key:       "33",
									Value:     pointerTo[float64](5),
									NotExists: 0,
									Ts:        timestamppb.New(testSomeMoment),
								},
								{
									Key:       "22",
									Value:     pointerTo[float64](8),
									NotExists: 1,
									Ts:        timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
								},
							},
							NotExists: 2,
						},
					},
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED,
						Message: "some error",
					},
				},
				StartedAt: timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
				ExpiresAt: timestamppb.New(testSomeMoment.Add(30 * time.Second)),
				Progress:  1,
				DiskUsage: 512,
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: testSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
				resp: &seqapi.FetchAsyncSearchResultResponse{
					Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
					Request: &seqapi.StartAsyncSearchRequest{
						Retention: durationpb.New(60 * time.Second),
						Query:     "message:error",
						From:      timestamppb.New(testSomeMoment.Add(-15 * time.Minute)),
						To:        timestamppb.New(testSomeMoment),
						Aggs: []*seqapi.AggregationQuery{
							{
								Field:     "x",
								GroupBy:   "level",
								Func:      seqapi.AggFunc_AGG_FUNC_AVG,
								Quantiles: []float64{0.9, 0.5},
							},
							{
								Field:    "y",
								GroupBy:  "level",
								Func:     seqapi.AggFunc_AGG_FUNC_SUM,
								Interval: pointerTo("30s"),
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
								Time: timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
							},
							{
								Id: "017a854298010000-8502fe7f2aa33df3",
								Data: map[string]string{
									"level":   "2",
									"message": "some error 2",
									"x":       "8",
								},
								Time: timestamppb.New(testSomeMoment.Add(-2 * time.Minute)),
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
										Value:     pointerTo[float64](2),
										NotExists: 0,
										Quantiles: []float64{2, 1},
									},
									{
										Key:       "2",
										Value:     pointerTo[float64](8),
										NotExists: 1,
										Quantiles: []float64{7, 4},
									},
								},
							},
							{
								Buckets: []*seqapi.Aggregation_Bucket{
									{
										Key:       "33",
										Value:     pointerTo[float64](2),
										NotExists: 0,
										Ts:        timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
									},
									{
										Key:       "33",
										Value:     pointerTo[float64](5),
										NotExists: 0,
										Ts:        timestamppb.New(testSomeMoment),
									},
									{
										Key:       "22",
										Value:     pointerTo[float64](8),
										NotExists: 1,
										Ts:        timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
									},
								},
								NotExists: 2,
							},
						},
						Error: &seqapi.Error{
							Code:    seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED,
							Message: "some error",
						},
					},
					StartedAt: timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
					ExpiresAt: timestamppb.New(testSomeMoment.Add(30 * time.Second)),
					Progress:  1,
					DiskUsage: 512,
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "partial_response",
			req: fetchAsyncSearchResultRequest{
				SearchID: testSearchID,
				Limit:    2,
				Offset:   10,
				Order:    oDESC,
			},
			want: fetchAsyncSearchResultResponseFromProto(&seqapi.FetchAsyncSearchResultResponse{
				Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
				Request: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     "message:error",
					From:      timestamppb.New(testSomeMoment.Add(-15 * time.Minute)),
					To:        timestamppb.New(testSomeMoment),
					WithDocs:  true,
					Size:      100,
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
							Time: timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
						},
						{
							Id: "017a854298010000-8502fe7f2aa33df3",
							Data: map[string]string{
								"level":   "2",
								"message": "some error 2",
								"x":       "8",
							},
							Time: timestamppb.New(testSomeMoment.Add(-2 * time.Minute)),
						},
					},
					Total: 2,
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED,
						Message: "some error",
					},
				},
				StartedAt: timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
				ExpiresAt: timestamppb.New(testSomeMoment.Add(30 * time.Second)),
				Progress:  1,
				DiskUsage: 512,
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
					Message: "partial response",
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: testSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
				resp: &seqapi.FetchAsyncSearchResultResponse{
					Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
					Request: &seqapi.StartAsyncSearchRequest{
						Retention: durationpb.New(60 * time.Second),
						Query:     "message:error",
						From:      timestamppb.New(testSomeMoment.Add(-15 * time.Minute)),
						To:        timestamppb.New(testSomeMoment),
						WithDocs:  true,
						Size:      100,
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
								Time: timestamppb.New(testSomeMoment.Add(-1 * time.Minute)),
							},
							{
								Id: "017a854298010000-8502fe7f2aa33df3",
								Data: map[string]string{
									"level":   "2",
									"message": "some error 2",
									"x":       "8",
								},
								Time: timestamppb.New(testSomeMoment.Add(-2 * time.Minute)),
							},
						},
						Total: 2,
						Error: &seqapi.Error{
							Code:    seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED,
							Message: "some error",
						},
					},
					StartedAt: timestamppb.New(testSomeMoment.Add(-30 * time.Second)),
					ExpiresAt: timestamppb.New(testSomeMoment.Add(30 * time.Second)),
					Progress:  1,
					DiskUsage: 512,
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
				},
			},
		},
		{
			name: "err_limit",
			req: fetchAsyncSearchResultRequest{
				SearchID: testSearchID,
				Limit:    -10,
				Offset:   20,
			},
			wantErr: true,
		},
		{
			name: "err_offset",
			req: fetchAsyncSearchResultRequest{
				SearchID: testSearchID,
				Limit:    10,
				Offset:   -20,
			},
			wantErr: true,
		},
		{
			name: "invalid_id",
			req: fetchAsyncSearchResultRequest{
				SearchID: "some invalid id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			svcMock := mock_asyncsearches.NewMockService(ctrl)

			seqData := test.APITestData{}
			seqData.Mocks.AsyncSearchesSvc = svcMock

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					FetchAsyncSearchResult(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[fetchAsyncSearchResultRequest, fetchAsyncSearchResultResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/async_search/fetch",
				Req:     tt.req,
				Handler: api.serveFetchAsyncSearchResult,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeFetchAsyncSearchResult_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[fetchAsyncSearchResultRequest, struct{}]{
		Method:  http.MethodPost,
		Target:  "/seqapi/v1/async_search/fetch",
		Handler: api.serveFetchAsyncSearchResult,
		WantErr: true,
	})
}

func pointerTo[T any](in T) *T {
	return &in
}
