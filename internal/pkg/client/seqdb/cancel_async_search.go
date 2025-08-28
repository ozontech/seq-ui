package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) CancelAsyncSearch(
	ctx context.Context,
	req *seqapi.CancelAsyncSearchRequest,
) (*seqapi.CancelAsyncSearchResponse, error) {
	proxyReq := newProxyCancelAsyncSearchRequest(req)
	_, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.CancelAsyncSearch(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	return &seqapi.CancelAsyncSearchResponse{}, nil
}
