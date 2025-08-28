package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *GRPCClient) GetFields(ctx context.Context, _ *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	proxyReq := &seqproxyapi.MappingRequest{}
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.Mapping(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}
	resp, err := (*proxyGetFieldsResp)(proxyResp.(*seqproxyapi.MappingResponse)).toProto()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}
