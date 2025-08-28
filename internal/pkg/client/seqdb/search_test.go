package seqdb

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_GRPCClient_Search(t *testing.T) {
	from := time.Now()
	to := from.Add(time.Second)
	var limit int32 = 3

	eventTime := timestamppb.New(time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC))

	events := make([]*seqapi.Event, limit)
	for i := 0; i < len(events); i++ {
		events[i] = makeEvent(fmt.Sprintf("test%d", i+1), i+1, eventTime)
	}

	docs := make([]*seqproxyapi.Document, 0, len(events))
	for i := 0; i < len(events); i++ {
		data, err := json.Marshal(events[i].Data)
		assert.NoError(t, err)
		docs = append(docs, &seqproxyapi.Document{
			Id:   events[i].Id,
			Data: data,
			Time: events[i].Time,
		})
	}

	type mockArgs struct {
		req  *seqproxyapi.ComplexSearchRequest
		resp *seqproxyapi.ComplexSearchResponse
		err  error
	}

	prepareMockArgs := func(req *seqapi.SearchRequest, docs []*seqproxyapi.Document, resp *seqapi.SearchResponse, err error) mockArgs {
		var proxyReq *seqproxyapi.ComplexSearchRequest
		var proxyResp *seqproxyapi.ComplexSearchResponse

		if req != nil {
			proxyReq = &seqproxyapi.ComplexSearchRequest{
				Query:     makeProxySearchQuery(req.Query, req.From, req.To),
				Size:      int64(req.Limit),
				Offset:    int64(req.Offset),
				WithTotal: req.WithTotal,
			}

			if req.Histogram != nil && req.Histogram.Interval != "" {
				proxyReq.Hist = &seqproxyapi.HistQuery{
					Interval: req.Histogram.Interval,
				}
			}
			if len(req.Aggregations) > 0 {
				proxyReq.Aggs = make([]*seqproxyapi.AggQuery, len(req.Aggregations))
				buf := make([]seqproxyapi.AggQuery, len(req.Aggregations))
				for i, agg := range req.Aggregations {
					q := &buf[i]
					newProxyAggQuery(agg, q)
					proxyReq.Aggs[i] = q
				}
			}
		}

		if resp != nil {
			proxyResp = &seqproxyapi.ComplexSearchResponse{
				Total: resp.Total,
			}

			proxyResp.Docs = docs
			if resp.Error != nil {
				proxyResp.Error = &seqproxyapi.Error{
					Code:    seqproxyapi.ErrorCode(resp.Error.Code),
					Message: resp.Error.Message,
				}
			}
			if resp.Histogram != nil {
				proxyResp.Hist = &seqproxyapi.Histogram{
					Buckets: make([]*seqproxyapi.Histogram_Bucket, 0, len(resp.Histogram.Buckets)),
				}
				for _, b := range resp.Histogram.Buckets {
					proxyResp.Hist.Buckets = append(proxyResp.Hist.Buckets, &seqproxyapi.Histogram_Bucket{
						DocCount: b.DocCount,
						Ts:       timestamppb.New(time.UnixMilli(int64(b.Key))),
					})
				}
			}
			for _, agg := range resp.Aggregations {
				proxyAgg := &seqproxyapi.Aggregation{
					Buckets:   make([]*seqproxyapi.Aggregation_Bucket, 0, len(agg.Buckets)),
					NotExists: agg.NotExists,
				}
				for _, b := range agg.Buckets {
					proxyAgg.Buckets = append(proxyAgg.Buckets, &seqproxyapi.Aggregation_Bucket{
						Key:       b.Key,
						Value:     *b.Value,
						NotExists: b.NotExists,
						Quantiles: b.Quantiles,
					})
				}
				proxyResp.Aggs = append(proxyResp.Aggs, proxyAgg)
			}
		}

		return mockArgs{
			req:  proxyReq,
			resp: proxyResp,
			err:  err,
		}
	}

	tests := []struct {
		name string

		req  *seqapi.SearchRequest
		docs []*seqproxyapi.Document

		wantResp *seqapi.SearchResponse
		wantErr  error
	}{
		{
			name: "ok",
			req: &seqapi.SearchRequest{
				Query:     "test_ok",
				From:      timestamppb.New(from),
				To:        timestamppb.New(to),
				Limit:     limit,
				Offset:    0,
				WithTotal: true,
			},
			docs: docs,
			wantResp: &seqapi.SearchResponse{
				Events:       events,
				Total:        int64(len(events)),
				Aggregations: makeAggregations(0, 0, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_histogram",
			req: &seqapi.SearchRequest{
				Query:  "test_histogram",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
				Histogram: &seqapi.SearchRequest_Histogram{
					Interval: "5s",
				},
			},
			docs: docs,
			wantResp: &seqapi.SearchResponse{
				Events:       events,
				Histogram:    makeHistogram(3),
				Aggregations: makeAggregations(0, 0, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_empty_histogram_interval",
			req: &seqapi.SearchRequest{
				Query:     "test_empty_hist_interval",
				From:      timestamppb.New(from),
				To:        timestamppb.New(to),
				Limit:     limit,
				Offset:    0,
				Histogram: &seqapi.SearchRequest_Histogram{},
			},
			docs: docs,
			wantResp: &seqapi.SearchResponse{
				Events:       events,
				Aggregations: makeAggregations(0, 0, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_aggs",
			req: &seqapi.SearchRequest{
				Query:  "test_aggs",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
				Aggregations: []*seqapi.AggregationQuery{
					{
						Field:   "test1",
						GroupBy: "service",
						Func:    seqapi.AggFunc_AGG_FUNC_MAX,
					},
				},
			},
			docs: docs,
			wantResp: &seqapi.SearchResponse{
				Events: events,
				Aggregations: makeAggregations(2, 3, &makeAggOpts{
					NotExists: 2,
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_invalid_utf8",
			req: &seqapi.SearchRequest{
				Query:  "test_invalid_utf8",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			docs: []*seqproxyapi.Document{
				{Id: "test1", Data: []byte("{\"key1\":\"val1\"}")},
				{Id: "test2", Data: []byte("{\"key1\":\"val1\",\"key2\":\"\xfdval\xff2\xfe\"}")},
			},
			wantResp: &seqapi.SearchResponse{
				Events: []*seqapi.Event{
					{Id: "test1", Data: map[string]string{"key1": "val1"}},
					{Id: "test2", Data: map[string]string{"key1": "val1", "key2": "�val�2�"}},
				},
				Aggregations: makeAggregations(0, 0, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "err_partial_response",
			req: &seqapi.SearchRequest{
				Query:  "test_partial_resp",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			docs: docs,
			wantResp: &seqapi.SearchResponse{
				Events:       events,
				Aggregations: makeAggregations(0, 0, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
				},
			},
		},
		{
			name: "err_proxy",
			req: &seqapi.SearchRequest{
				Query:  "test_err_proxy",
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Limit:  limit,
				Offset: 0,
			},
			wantErr: errors.New("proxy error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mArgs := prepareMockArgs(tt.req, tt.docs, tt.wantResp, tt.wantErr)

			ctrl := gomock.NewController(t)
			seqProxyMock := mock.NewMockSeqProxyApiClient(ctrl)
			seqProxyMock.EXPECT().ComplexSearch(ctx, mArgs.req).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			resp, err := c.Search(ctx, tt.req)

			require.Equal(t, tt.wantErr, err)
			if tt.wantResp == nil {
				require.Nil(t, resp)
			} else {
				require.Equal(t, tt.wantResp.Total, resp.Total)
				require.Equal(t, tt.wantResp.Error, resp.Error)

				require.Equal(t, len(tt.wantResp.Events), len(resp.Events))
				for i, e := range tt.wantResp.Events {
					require.True(t, checkEventsEqual(e, resp.Events[i]))
				}

				require.Equal(t, tt.wantResp.Histogram, resp.Histogram)
				require.Equal(t, tt.wantResp.Aggregations, resp.Aggregations)
			}
		})
	}
}
