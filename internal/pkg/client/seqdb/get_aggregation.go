package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) GetAggregation(ctx context.Context, req *seqapi.GetAggregationRequest) (*seqapi.GetAggregationResponse, error) {
	proxyReq := newProxyGetAggReq(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.GetAggregation(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp := (*proxyGetAggResp)(proxyResp.(*seqproxyapi.GetAggregationResponse)).toProto()
	return resp, nil
}
