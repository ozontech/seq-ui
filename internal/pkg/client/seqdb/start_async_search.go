package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) StartAsyncSearch(
	ctx context.Context,
	req *seqapi.StartAsyncSearchRequest,
) (*seqapi.StartAsyncSearchResponse, error) {
	proxyReq := newProxyStartAsyncSearchRequest(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.StartAsyncSearch(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp := (*proxyStartAsyncSearchResp)(proxyResp.(*seqproxyapi.StartAsyncSearchResponse)).toProto()
	return resp, nil
}
