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
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServeSearch(t *testing.T) {
	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)
	eventTime := from.Add(time.Millisecond) // 2023-09-25T10:20:30.001Z

	formatReqBody := func(limit, offset int, withTotal bool, histInterval string, aggQueries aggregationQueries, order string) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"limit":%d,"offset":%d`,
			query, from.Format(time.RFC3339), to.Format(time.RFC3339), limit, offset))

		if withTotal {
			sb.WriteString(fmt.Sprintf(`,"withTotal":%v`, withTotal))
		}
		if histInterval != "" {
			sb.WriteString(fmt.Sprintf(`,"histogram":{"interval":%q}`, histInterval))
		}
		if len(aggQueries) > 0 {
			aggQueriesRaw, err := json.Marshal(aggQueries)
			assert.NoError(t, err)
			sb.WriteString(fmt.Sprintf(`,"aggregations":%s`, aggQueriesRaw))
		}
		if order != "" {
			sb.WriteString(fmt.Sprintf(`,"order":%q`, order))
		}

		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req  *seqapi.SearchRequest
		resp *seqapi.SearchResponse
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
			name:    "ok_simple",
			reqBody: formatReqBody(3, 0, false, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test3","data":{"field1":"val1","field2":"val2","field3":"val3"},"time":"2023-09-25T10:20:30.001Z"}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name:    "ok_with_total",
			reqBody: formatReqBody(3, 0, true, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:     query,
					From:      timestamppb.New(from),
					To:        timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test3","data":{"field1":"val1","field2":"val2","field3":"val3"},"time":"2023-09-25T10:20:30.001Z"}],"total":"10","error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name:    "ok_order_asc",
			reqBody: formatReqBody(3, 0, false, "", nil, string(oASC)),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test3","data":{"field1":"val1","field2":"val2","field3":"val3"},"time":"2023-09-25T10:20:30.001Z"}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name:    "ok_with_hist",
			reqBody: formatReqBody(3, 0, false, "5s", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test3","data":{"field1":"val1","field2":"val2","field3":"val3"},"time":"2023-09-25T10:20:30.001Z"}],"histogram":{"buckets":[{"key":"0","docCount":"1"},{"key":"100","docCount":"2"}]},"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name: "ok_with_aggs",
			reqBody: formatReqBody(3, 0, false, "", aggregationQueries{
				{
					Field:   "test1",
					GroupBy: "service",
					Func:    afAvg,
				},
			}, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2023-09-25T10:20:30.001Z"},{"id":"test3","data":{"field1":"val1","field2":"val2","field3":"val3"},"time":"2023-09-25T10:20:30.001Z"}],"aggregations":[{"buckets":[{"key":"test1","value":1,"not_exists":5},{"key":"test2","value":2,"not_exists":5},{"key":"test3","value":3,"not_exists":5}]},{"buckets":[{"key":"test1","value":1,"not_exists":5},{"key":"test2","value":2,"not_exists":5},{"key":"test3","value":3,"not_exists":5}]}],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit:            5,
				MaxAggregationsPerRequest: 5,
			}),
		},
		{
			name:    "ok_empty_events",
			reqBody: formatReqBody(3, 0, false, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[],"error":{"code":"ERROR_CODE_NO"},"partialResponse":false}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name:    "err_partial_response",
			reqBody: formatReqBody(3, 0, false, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
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
			wantRespBody: `{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"}],"error":{"code":"ERROR_CODE_PARTIAL_RESPONSE","message":"partial response"},"partialResponse":true}`,
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_search_limit_zero",
			reqBody:    formatReqBody(0, 0, false, "", nil, ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_search_limit_max",
			reqBody:    formatReqBody(10, 0, false, "", nil, ""),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				MaxSearchLimit: 5,
			},
		},
		{
			name:       "err_aggs_limit_max",
			reqBody:    formatReqBody(3, 0, false, "", aggregationQueries{{}, {}, {}}, ""),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				MaxSearchLimit:            5,
				MaxAggregationsPerRequest: 2,
			},
		}, {
			name:       "err_offset_too_high",
			reqBody:    formatReqBody(3, 11, false, "", nil, ""),
			wantStatus: http.StatusBadRequest,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit:       5,
				MaxSearchOffsetLimit: 10,
			}),
		},
		{
			name:    "err_total_too_high",
			reqBody: formatReqBody(3, 0, true, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:     query,
					From:      timestamppb.New(from),
					To:        timestamppb.New(to),
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
			wantRespBody: fmt.Sprintf(`{"events":[{"id":"test1","data":{"field1":"val1"},"time":"2023-09-25T10:20:30.001Z"}],"total":"11","error":{"code":"ERROR_CODE_QUERY_TOO_HEAVY","message":%q},"partialResponse":false}`, api_error.ErrQueryTooHeavy.Error()),
			wantStatus:   http.StatusOK,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit:      5,
				MaxSearchTotalLimit: 10,
			}),
		},
		{
			name:    "err_client",
			reqBody: formatReqBody(3, 0, false, "", nil, ""),
			mockArgs: &mockArgs{
				req: &seqapi.SearchRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  3,
					Offset: 0,
				},
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
			cfg: test.SetCfgDefaults(config.SeqAPI{
				MaxSearchLimit: 5,
			}),
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
				seqDbMock.EXPECT().Search(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/search", strings.NewReader(tt.reqBody))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveSearch,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
