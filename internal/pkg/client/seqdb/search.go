package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *GRPCClient) Search(ctx context.Context, req *seqapi.SearchRequest) (*seqapi.SearchResponse, error) {
	proxyReq := newProxySearchReq(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.ComplexSearch(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp, err := (*proxySearchResp)(proxyResp.(*seqproxyapi.ComplexSearchResponse)).toProto()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
