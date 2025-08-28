package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) GetHistogram(ctx context.Context, req *seqapi.GetHistogramRequest) (*seqapi.GetHistogramResponse, error) {
	proxyReq := newProxyGetHistReq(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.GetHistogram(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp := (*proxyGetHistResp)(proxyResp.(*seqproxyapi.GetHistogramResponse)).toProto()
	return resp, nil
}
