package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) GetAsyncSearchesList(
	ctx context.Context,
	req *seqapi.GetAsyncSearchesListRequest,
	ids []string,
) (*seqapi.GetAsyncSearchesListResponse, error) {
	proxyReq := newProxyGetAsyncSearchesListRequest(req, ids)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.GetAsyncSearchesList(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp := (*proxyGetAsyncSearchesListResp)(proxyResp.(*seqproxyapi.GetAsyncSearchesListResponse)).toProto()
	return resp, nil
}
