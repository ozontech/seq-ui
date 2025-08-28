package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) Status(ctx context.Context, _ *seqapi.StatusRequest) (*seqapi.StatusResponse, error) {
	proxyReq := &seqproxyapi.StatusRequest{}
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.Status(ctx, proxyReq)
		},
	)
	if err != nil {
		return nil, err
	}

	resp := (*proxyStatusResp)(proxyResp.(*seqproxyapi.StatusResponse)).toProto()
	return resp, nil
}
