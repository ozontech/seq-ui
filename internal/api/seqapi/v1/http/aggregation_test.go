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

func TestServeGetAggregation(t *testing.T) {
	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)

	formatReqBody := func(aggField string, aggQueries aggregationQueries) string {
		if len(aggQueries) > 0 {
			aggQueriesRaw, err := json.Marshal(aggQueries)
			assert.NoError(t, err)
			return fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"aggField":%q,"aggregations":%s}`,
				query, from.Format(time.RFC3339), to.Format(time.RFC3339), aggField, aggQueriesRaw)
		}
		return fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"aggField":%q}`,
			query, from.Format(time.RFC3339), to.Format(time.RFC3339), aggField)
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
			name:    "ok_single_agg",
			reqBody: formatReqBody("test_single", nil),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
					AggField: "test_single",
				},
				resp: &seqapi.GetAggregationResponse{
					Aggregation:  test.MakeAggregation(2, nil),
					Aggregations: test.MakeAggregations(1, 2, nil),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"aggregation":{"buckets":[{"key":"test1","value":1},{"key":"test2","value":2}]},"aggregations":[{"buckets":[{"key":"test1","value":1},{"key":"test2","value":2}]}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "ok_multi_agg",
			reqBody: formatReqBody("", aggregationQueries{
				{Field: "test_multi1"},
				{Field: "test_multi2"},
				{Field: "test_multi3"},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
					Aggregations: []*seqapi.AggregationQuery{
						{Field: "test_multi1"},
						{Field: "test_multi2"},
						{Field: "test_multi3"},
					},
				},
				resp: &seqapi.GetAggregationResponse{
					Aggregation:  test.MakeAggregation(3, nil),
					Aggregations: test.MakeAggregations(2, 3, nil),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"aggregation":{"buckets":[{"key":"test1","value":1},{"key":"test2","value":2},{"key":"test3","value":3}]},"aggregations":[{"buckets":[{"key":"test1","value":1},{"key":"test2","value":2},{"key":"test3","value":3}]},{"buckets":[{"key":"test1","value":1},{"key":"test2","value":2},{"key":"test3","value":3}]}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},

		{
			name: "ok_agg_quantile",
			reqBody: formatReqBody("", aggregationQueries{
				{
					Field:     "test_multi1",
					GroupBy:   "service",
					Func:      afQuantile,
					Quantiles: []float64{0.95, 0.99},
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: query,
					From:  timestamppb.New(from),
					To:    timestamppb.New(to),
					Aggregations: []*seqapi.AggregationQuery{
						{
							Field:     "test_multi1",
							GroupBy:   "service",
							Func:      seqapi.AggFunc_AGG_FUNC_QUANTILE,
							Quantiles: []float64{0.95, 0.99},
						},
					},
				},
				resp: &seqapi.GetAggregationResponse{
					Aggregation: test.MakeAggregation(3, &test.MakeAggOpts{
						NotExists: 10,
						Quantiles: []float64{100, 150},
					}),
					Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
						NotExists: 10,
						Quantiles: []float64{100, 150},
					}),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
			wantRespBody: `{"aggregation":{"buckets":[{"key":"test1","value":1,"not_exists":10,"quantiles":[100,150]},{"key":"test2","value":2,"not_exists":10,"quantiles":[100,150]},{"key":"test3","value":3,"not_exists":10,"quantiles":[100,150]}]},"aggregations":[{"buckets":[{"key":"test1","value":1,"not_exists":10,"quantiles":[100,150]},{"key":"test2","value":2,"not_exists":10,"quantiles":[100,150]},{"key":"test3","value":3,"not_exists":10,"quantiles":[100,150]}]},{"buckets":[{"key":"test1","value":1,"not_exists":10,"quantiles":[100,150]},{"key":"test2","value":2,"not_exists":10,"quantiles":[100,150]},{"key":"test3","value":3,"not_exists":10,"quantiles":[100,150]}]}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name:    "err_partial_response",
			reqBody: formatReqBody("test_err_partial", nil),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
					AggField: "test_err_partial",
				},
				resp: &seqapi.GetAggregationResponse{
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
					PartialResponse: true,
				},
			},
			wantRespBody: `{"aggregation":{"buckets":[]},"aggregations":[],"error":{"code":"ERROR_CODE_PARTIAL_RESPONSE","message":"partial response"},"partialResponse":true}`,
			wantStatus:   http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_aggs_limit_max",
			reqBody:    formatReqBody("", aggregationQueries{{}, {}, {}}),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 2,
				},
			},
		},
		{
			name:    "err_client",
			reqBody: formatReqBody("test_err_client", nil),
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    query,
					From:     timestamppb.New(from),
					To:       timestamppb.New(to),
					AggField: "test_err_client",
				},
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
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
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/aggregation", strings.NewReader(tt.reqBody))
			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetAggregation,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
