package seqdb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) FetchAsyncSearchResult(
	ctx context.Context,
	req *seqapi.FetchAsyncSearchResultRequest,
) (*seqapi.FetchAsyncSearchResultResponse, error) {
	proxyReq := newProxyFetchAsyncSearchResultRequest(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.FetchAsyncSearchResult(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp, err := (*proxyFetchAsyncSearchResultResp)(proxyResp.(*seqproxyapi.FetchAsyncSearchResultResponse)).toProto()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
