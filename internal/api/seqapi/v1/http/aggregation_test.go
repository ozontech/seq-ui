package http

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetAggregation(t *testing.T) {
	type mockArgs struct {
		req  *seqapi.GetAggregationRequest
		resp *seqapi.GetAggregationResponse
		err  error
	}

	tests := []struct {
		name string

		req     getAggregationRequest
		want    getAggregationResponse
		wantErr bool

		mockArgs *mockArgs
		cfg      config.SeqAPI
	}{
		{
			name: "ok_single_agg",
			req: getAggregationRequest{
				Query:    testQuery,
				From:     testFrom,
				To:       testTo,
				AggField: "test_single",
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    testQuery,
					From:     timestamppb.New(testFrom),
					To:       timestamppb.New(testTo),
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
			want: getAggregationResponse{
				Aggregation:     aggregationFromProto(test.MakeAggregation(2, nil)),
				Aggregations:    aggregationsFromProto(test.MakeAggregations(1, 2, nil), true),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "ok_multi_agg",
			req: getAggregationRequest{
				Query: testQuery,
				From:  testFrom,
				To:    testTo,
				Aggregations: aggregationQueries{
					{Field: "test_multi1"},
					{Field: "test_multi2"},
					{Field: "test_multi3"},
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: testQuery,
					From:  timestamppb.New(testFrom),
					To:    timestamppb.New(testTo),
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
			want: getAggregationResponse{
				Aggregation:     aggregationFromProto(test.MakeAggregation(3, nil)),
				Aggregations:    aggregationsFromProto(test.MakeAggregations(2, 3, nil), true),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},

		{
			name: "ok_agg_quantile",
			req: getAggregationRequest{
				Query: testQuery,
				From:  testFrom,
				To:    testTo,
				Aggregations: aggregationQueries{
					{
						Field:     "test_multi1",
						GroupBy:   "service",
						Func:      afQuantile,
						Quantiles: []float64{0.95, 0.99},
					},
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query: testQuery,
					From:  timestamppb.New(testFrom),
					To:    timestamppb.New(testTo),
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
			want: getAggregationResponse{
				Aggregation: aggregationFromProto(test.MakeAggregation(3, &test.MakeAggOpts{
					NotExists: 10,
					Quantiles: []float64{100, 150},
				})),
				Aggregations: aggregationsFromProto(test.MakeAggregations(2, 3, &test.MakeAggOpts{
					NotExists: 10,
					Quantiles: []float64{100, 150},
				}), true),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "err_partial_response",
			req: getAggregationRequest{
				Query:    testQuery,
				From:     testFrom,
				To:       testTo,
				AggField: "test_err_partial",
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    testQuery,
					From:     timestamppb.New(testFrom),
					To:       timestamppb.New(testTo),
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
			want: getAggregationResponse{
				Aggregation:     aggregationFromProto(nil),
				Aggregations:    aggregationsFromProto(nil, true),
				Error:           apiError{Code: aecPartialResponse, Message: "partial response"},
				PartialResponse: true,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "err_aggs_limit_max",
			req: getAggregationRequest{
				Query:        testQuery,
				From:         testFrom,
				To:           testTo,
				Aggregations: aggregationQueries{{}, {}, {}},
			},
			wantErr: true,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 2,
				},
			},
		},
		{
			name: "err_client",
			req: getAggregationRequest{
				Query:    testQuery,
				From:     testFrom,
				To:       testTo,
				AggField: "test_err_client",
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAggregationRequest{
					Query:    testQuery,
					From:     timestamppb.New(testFrom),
					To:       timestamppb.New(testTo),
					AggField: "test_err_client",
				},
				err: errors.New("client error"),
			},
			wantErr: true,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
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

				seqDbMock.EXPECT().
					GetAggregation(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getAggregationRequest, getAggregationResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/aggregation",
				Req:     tt.req,
				Handler: api.serveGetAggregation,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
