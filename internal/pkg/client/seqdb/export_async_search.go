package seqdb

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (c *GRPCClient) ExportAsyncSearch(ctx context.Context, req *seqapi.ExportAsyncSearchRequest, cw *httputil.ChunkedWriter) error {
	proxyReq := newProxyExportAsyncSearchReq(req)
	proxyResp, err := c.sendRequest(ctx,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.ExportAsyncSearch(ctx, proxyReq)
		},
	)
	if err != nil {
		return err
	}

	stream := proxyResp.(seqproxyapi.SeqProxyApi_ExportAsyncSearchClient)
	return c.writeExportStream(stream, req.Format, req.Fields, cw, "ExportAsyncSearch")
}
