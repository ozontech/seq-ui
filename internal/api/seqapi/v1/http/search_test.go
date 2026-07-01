package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeSearch(t *testing.T) {
	var (
		eventTime = testTimestamp.Add(time.Millisecond)
	)

	type mockArgs struct {
		req  *seqapi.SearchRequest
		resp *seqapi.SearchResponse
		err  error
	}

	tests := []struct {
		name string

		req     searchRequest
		want    searchResponse
		cfg     config.SeqAPI
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_simple",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(3, eventTime)),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(3, eventTime),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_with_total",
			req: searchRequest{
				Query:     testQuery,
				From:      testTimestamp,
				To:        testTimestamp.Add(time.Second),
				Limit:     3,
				WithTotal: true,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(3, eventTime)),
				Total:           "10",
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:     testQuery,
					From:      timestamppb.New(testTimestamp),
					To:        timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:     3,
					Offset:    0,
					WithTotal: true,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(3, eventTime),
					Total:  10,
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_order_asc",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
				Order: oASC,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(3, eventTime)),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
					Order:  seqapi.Order_ORDER_ASC,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(3, eventTime),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_with_hist",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
				Histogram: struct {
					Interval string "json:\"interval\""
				}{Interval: "5s"},
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(3, eventTime)),
				Histogram:       histogramFromProto(test.MakeHistogram(2), false),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
					Histogram: &seqapi.SearchRequest_Histogram{
						Interval: "5s",
					},
				},
				resp: &seqapi.SearchResponse{
					Events:    test.MakeEvents(3, eventTime),
					Histogram: test.MakeHistogram(2),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_with_aggs",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
				Aggregations: aggregationQueries{
					{
						Field:   "test1",
						GroupBy: "service",
						Func:    afAvg,
					},
				},
			},
			want: searchResponse{
				Events: eventsFromProto(test.MakeEvents(3, eventTime)),
				Aggregations: aggregationsFromProto(test.MakeAggregations(2, 3, &test.MakeAggOpts{
					NotExists: 5,
				}), false),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
					Aggregations: []*seqapi.AggregationQuery{
						{
							Field:   "test1",
							GroupBy: "service",
							Func:    seqapi.AggFunc_AGG_FUNC_AVG,
						},
					},
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(3, eventTime),
					Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
						NotExists: 5,
					}),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "ok_empty_events",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(0, eventTime)),
				Error:           apiError{Code: aecNo},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(0, eventTime),
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
			},
		},
		{
			name: "err_partial_response",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(1, eventTime)),
				Error:           apiError{Code: aecPartialResponse, Message: "partial response"},
				PartialResponse: true,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(1, eventTime),
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
						Message: "partial response",
					},
					PartialResponse: true,
				},
			},
		},
		{
			name: "err_search_limit_zero",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 0,
			},
			wantErr: true,
		},
		{
			name: "err_search_limit_max",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 10,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			},
			wantErr: true,
		},
		{
			name: "err_aggs_limit_max",
			req: searchRequest{
				Query:        testQuery,
				From:         testTimestamp,
				To:           testTimestamp.Add(time.Second),
				Limit:        3,
				Aggregations: aggregationQueries{{}, {}, {}},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:            5,
					MaxAggregationsPerRequest: 2,
				},
			},
			wantErr: true,
		},
		{
			name: "err_offset_too_high",
			req: searchRequest{
				Query:  testQuery,
				From:   testTimestamp,
				To:     testTimestamp.Add(time.Second),
				Limit:  3,
				Offset: 11,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:       5,
					MaxSearchOffsetLimit: 10,
				},
			}),
			wantErr: true,
		},
		{
			name: "err_total_too_high",
			req: searchRequest{
				Query:     testQuery,
				From:      testTimestamp,
				To:        testTimestamp.Add(time.Second),
				Limit:     3,
				WithTotal: true,
			},
			want: searchResponse{
				Events:          eventsFromProto(test.MakeEvents(1, eventTime)),
				Total:           "11",
				Error:           apiError{Code: aecQueryTooHeavy, Message: api_error.ErrQueryTooHeavy.Error()},
				PartialResponse: false,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:       5,
					MaxSearchOffsetLimit: 10,
				},
			}),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:     testQuery,
					From:      timestamppb.New(testTimestamp),
					To:        timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:     3,
					Offset:    0,
					WithTotal: true,
				},
				resp: &seqapi.SearchResponse{
					Events: test.MakeEvents(1, eventTime),
					Total:  11,
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY,
						Message: api_error.ErrQueryTooHeavy.Error(),
					},
				},
			},
		},
		{
			name: "err_client",
			req: searchRequest{
				Query: testQuery,
				From:  testTimestamp,
				To:    testTimestamp.Add(time.Second),
				Limit: 3,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			wantErr: true,
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  testQuery,
					From:   timestamppb.New(testTimestamp),
					To:     timestamppb.New(testTimestamp.Add(time.Second)),
					Limit:  3,
					Offset: 0,
				},
				err: errSomethingWrong,
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
					Search(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[searchRequest, searchResponse]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/search",
				Req:     tt.req,
				Handler: api.serveSearch,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
