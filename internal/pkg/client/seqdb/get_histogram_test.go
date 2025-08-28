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

func Test_GRPCClient_GetHistogram(t *testing.T) {
	from := time.Now()
	to := from.Add(time.Second)

	type mockArgs struct {
		req  *seqproxyapi.GetHistogramRequest
		resp *seqproxyapi.GetHistogramResponse
		err  error
	}

	prepareMockArgs := func(req *seqapi.GetHistogramRequest, resp *seqapi.GetHistogramResponse, err error) mockArgs {
		var proxyReq *seqproxyapi.GetHistogramRequest
		var proxyResp *seqproxyapi.GetHistogramResponse

		if req != nil {
			proxyReq = &seqproxyapi.GetHistogramRequest{
				Query: makeProxySearchQuery(req.Query, req.From, req.To),
				Hist: &seqproxyapi.HistQuery{
					Interval: req.Interval,
				},
			}
		}

		if resp != nil {
			proxyResp = &seqproxyapi.GetHistogramResponse{}

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
		}

		return mockArgs{
			req:  proxyReq,
			resp: proxyResp,
			err:  err,
		}
	}

	tests := []struct {
		name string

		req      *seqapi.GetHistogramRequest
		wantResp *seqapi.GetHistogramResponse
		wantErr  error
	}{
		{
			name: "ok",
			req: &seqapi.GetHistogramRequest{
				Query:    "test_ok",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				Interval: "5s",
			},
			wantResp: &seqapi.GetHistogramResponse{
				Histogram: makeHistogram(3),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "ok_partial_response",
			req: &seqapi.GetHistogramRequest{
				Query:    "test_ok_partial_resp",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				Interval: "5s",
			},
			wantResp: &seqapi.GetHistogramResponse{
				Histogram: makeHistogram(3),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
				},
			},
		},
		{
			name: "err_proxy",
			req: &seqapi.GetHistogramRequest{
				Query:    "test_err_proxy",
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				Interval: "5s",
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
			seqProxyMock.EXPECT().GetHistogram(ctx, mArgs.req).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			resp, err := c.GetHistogram(ctx, tt.req)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.wantResp, resp)
		})
	}
}
