package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) DeleteAsyncSearch(
	ctx context.Context,
	req *seqapi.DeleteAsyncSearchRequest,
) (*seqapi.DeleteAsyncSearchResponse, error) {
	proxyReq := newProxyDeleteAsyncSearchRequest(req)
	_, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.DeleteAsyncSearch(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	return &seqapi.DeleteAsyncSearchResponse{}, nil
}
