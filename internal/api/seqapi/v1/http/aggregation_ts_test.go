package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServeGetAggregationTs(t *testing.T) {
	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(5 * time.Second)
	interval := "1s"

	formatReqBody := func(aggQueries aggregationTsQueries) string {
		aggQueriesRaw, err := json.Marshal(aggQueries)
		assert.NoError(t, err)
		return fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"aggregations":%s}`,
			query, from.Format(time.RFC3339), to.Format(time.RFC3339), aggQueriesRaw)
	}

	type mockArgs struct {
		req  *seqapi.GetAggregationRequest
		resp *seqapi.GetAggregationResponse
		err  error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
		cfg      config.SeqAPI
	}{
		{
			name: "ok_count",
			reqBody: formatReqBody(aggregationTsQueries{
				{
					aggregationQuery: aggregationQuery{
						Field: "test_count1",
						Func:  afCount,
					},
					Interval: interval,
				},
				{
					aggregationQuery: aggregationQuery{
						Field: "test_count2",
						Func:  afCount,
					},
					Interval: interval,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
					Aggregations: []*seqapi.AggregationQuery{
						{Field: "test_count1", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval},
						{Field: "test_count2", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval},
					},
				},
				resp: &seqapi.GetAggregationResponse{
					Aggregation: test.MakeAggregation(3, &test.MakeAggOpts{
						Ts: []*timestamppb.Timestamp{
							timestamppb.New(from.Add(time.Second)),
							timestamppb.New(from.Add(2 * time.Second)),
							timestamppb.New(from.Add(3 * time.Second)),
						},
					}),
					Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
						Ts: []*timestamppb.Timestamp{
							timestamppb.New(from.Add(time.Second)),
							timestamppb.New(from.Add(2 * time.Second)),
							timestamppb.New(from.Add(3 * time.Second)),
						},
					}),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"aggregations":[{"data":{"result":[{"metric":{"test_count1":"test1"},"values":[{"timestamp":1695637231,"value":1}]},{"metric":{"test_count1":"test2"},"values":[{"timestamp":1695637232,"value":2}]},{"metric":{"test_count1":"test3"},"values":[{"timestamp":1695637233,"value":3}]}]}},{"data":{"result":[{"metric":{"test_count2":"test1"},"values":[{"timestamp":1695637231,"value":1}]},{"metric":{"test_count2":"test2"},"values":[{"timestamp":1695637232,"value":2}]},{"metric":{"test_count2":"test3"},"values":[{"timestamp":1695637233,"value":3}]}]}}]}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:  3,
				MaxBucketsPerAggregationTs: 100,
			},
		},
		{
			name: "ok_quantile",
			reqBody: formatReqBody(aggregationTsQueries{
				{
					aggregationQuery: aggregationQuery{
						Field:     "test_quantile1",
						GroupBy:   "service",
						Func:      afQuantile,
						Quantiles: []float64{0.95, 0.99},
					},
					Interval: interval,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
					Aggregations: []*seqapi.AggregationQuery{
						{
							Field:     "test_quantile1",
							GroupBy:   "service",
							Func:      seqapi.AggFunc_AGG_FUNC_QUANTILE,
							Quantiles: []float64{0.95, 0.99},
							Interval:  &interval,
						},
					},
				},
				resp: &seqapi.GetAggregationResponse{
					Aggregation: test.MakeAggregation(3, &test.MakeAggOpts{
						Quantiles: []float64{100, 150},
						Ts: []*timestamppb.Timestamp{
							timestamppb.New(from.Add(time.Second)),
							timestamppb.New(from.Add(2 * time.Second)),
							timestamppb.New(from.Add(3 * time.Second)),
						},
					}),
					Aggregations: test.MakeAggregations(1, 3, &test.MakeAggOpts{
						Quantiles: []float64{100, 150},
						Ts: []*timestamppb.Timestamp{
							timestamppb.New(from.Add(time.Second)),
							timestamppb.New(from.Add(2 * time.Second)),
							timestamppb.New(from.Add(3 * time.Second)),
						},
					}),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"aggregations":[{"data":{"result":[{"metric":{"quantile":"p95","service":"test1"},"values":[{"timestamp":1695637231,"value":100}]},{"metric":{"quantile":"p99","service":"test1"},"values":[{"timestamp":1695637231,"value":150}]},{"metric":{"quantile":"p95","service":"test2"},"values":[{"timestamp":1695637232,"value":100}]},{"metric":{"quantile":"p99","service":"test2"},"values":[{"timestamp":1695637232,"value":150}]},{"metric":{"quantile":"p95","service":"test3"},"values":[{"timestamp":1695637233,"value":100}]},{"metric":{"quantile":"p99","service":"test3"},"values":[{"timestamp":1695637233,"value":150}]}]}}]}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:  3,
				MaxBucketsPerAggregationTs: 100,
			},
		},
		{
			name:    "err_partial_response",
			reqBody: formatReqBody(nil),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
				},
				resp: &seqapi.GetAggregationResponse{
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
					PartialResponse: true,
				},
			},
			wantRespBody: `{"aggregations":null,"error":"partial response"}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:  3,
				MaxBucketsPerAggregationTs: 100,
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_aggs_limit_max",
			reqBody:    formatReqBody(aggregationTsQueries{{}, {}, {}}),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:  2,
				MaxBucketsPerAggregationTs: 100,
			},
		},
		{
			name: "err_buckets_limit_max",
			reqBody: formatReqBody(aggregationTsQueries{
				{
					Interval: "500ms",
				},
			}),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:  3,
				MaxBucketsPerAggregationTs: 8,
			},
		},
		{
			name:    "err_client",
			reqBody: formatReqBody(nil),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
				},
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAggregation(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/aggregation_ts", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetAggregationTs,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
