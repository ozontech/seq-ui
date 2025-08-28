package seqdb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_GRPCClient_GetAggregation(t *testing.T) {
	from := time.Now()
	to := from.Add(time.Second)

	type mockArgs struct {
		req  *seqproxyapi.GetAggregationRequest
		resp *seqproxyapi.GetAggregationResponse
		err  error
	}

	prepareMockArgs := func(req *seqapi.GetAggregationRequest, resp *seqapi.GetAggregationResponse, err error) mockArgs {
		var proxyReq *seqproxyapi.GetAggregationRequest
		var proxyResp *seqproxyapi.GetAggregationResponse

		if req != nil {
			proxyReq = &seqproxyapi.GetAggregationRequest{
				Query: makeProxySearchQuery(req.Query, req.From, req.To),
			}

			if len(req.Aggregations) == 0 && req.AggField != "" {
				proxyReq.Aggs = append(proxyReq.Aggs, &seqproxyapi.AggQuery{
					Field: req.AggField,
				})
			} else {
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
			proxyResp = &seqproxyapi.GetAggregationResponse{}

			if resp.Error != nil {
				proxyResp.Error = &seqproxyapi.Error{
					Code:    seqproxyapi.ErrorCode(resp.Error.Code),
					Message: resp.Error.Message,
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

		req      *seqapi.GetAggregationRequest
		wantResp *seqapi.GetAggregationResponse
		wantErr  error
	}{
		{
			name: "ok_single_agg",
			req: &seqapi.GetAggregationRequest{
				Query:    "test_ok_single_agg",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				AggField: "test1",
			},
			wantResp: &seqapi.GetAggregationResponse{
				Aggregation:  makeAggregation(2, nil),
				Aggregations: makeAggregations(1, 2, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_multi_agg",
			req: &seqapi.GetAggregationRequest{
				Query: "test_ok_multi_agg",
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
				},
			},
			wantResp: &seqapi.GetAggregationResponse{
				Aggregation:  makeAggregation(3, nil),
				Aggregations: makeAggregations(2, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_agg_quantile",
			req: &seqapi.GetAggregationRequest{
				Query: "test_ok_agg_quantile",
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{
						Field:     "test1",
						GroupBy:   "service",
						Func:      seqapi.AggFunc_AGG_FUNC_QUANTILE,
						Quantiles: []float64{0.95, 0.99},
					},
				},
			},
			wantResp: &seqapi.GetAggregationResponse{
				Aggregation: makeAggregation(3, &makeAggOpts{
					NotExists: 10,
					Quantiles: []float64{100, 150},
				}),
				Aggregations: makeAggregations(2, 3, &makeAggOpts{
					NotExists: 10,
					Quantiles: []float64{100, 150},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_partial_response",
			req: &seqapi.GetAggregationRequest{
				Query:    "test_ok_partial_resp",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				AggField: "test1",
			},
			wantResp: &seqapi.GetAggregationResponse{
				Aggregation:  makeAggregation(3, nil),
				Aggregations: makeAggregations(1, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
				},
			},
		},
		{
			name: "err_proxy",
			req: &seqapi.GetAggregationRequest{
				Query:    "test_err_proxy",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				AggField: "test1",
			},
			wantErr: errors.New("proxy error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mArgs := prepareMockArgs(tt.req, tt.wantResp, tt.wantErr)

			ctrl := gomock.NewController(t)
			seqProxyMock := mock.NewMockSeqProxyApiClient(ctrl)
			seqProxyMock.EXPECT().GetAggregation(ctx, mArgs.req).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			resp, err := c.GetAggregation(ctx, tt.req)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.wantResp, resp)
		})
	}
}
