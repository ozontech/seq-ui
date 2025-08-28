package seqdb

import (
	"context"
	"errors"
	"io"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *GRPCClient) GetEvent(ctx context.Context, req *seqapi.GetEventRequest) (*seqapi.GetEventResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	proxyReq := newProxyGetEventReq(req)
	proxyResp, err := c.sendRequest(ctxWithTimeout,
		func(client seqproxyapi.SeqProxyApiClient) (any, error) {
			return client.Fetch(ctxWithTimeout, proxyReq)
		},
	)
	out := proxyResp.(seqproxyapi.SeqProxyApi_FetchClient)
	logger.Debug("GetEvent.Fetch request results", zap.Any("out", out), zap.Error(err))
	if err != nil {
		return nil, err
	}
	docs := make([]*seqproxyapi.Document, 0)
	for {
		o, err := out.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			logger.Error("grpc client stream recv failed",
				zap.String("method", "Fetch"),
				zap.Error(err),
			)
			metric.SeqDBClientStreamError.WithLabelValues("Fetch", "recv").Inc()
			return nil, err
		}
		docs = append(docs, o)
	}

	event := &seqapi.Event{}
	if len(docs) > 0 {
		// false positive problem https://github.com/securego/gosec/issues/1005
		//nolint:gosec
		if err := (*proxyDoc)(docs[0]).toProto(event); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &seqapi.GetEventResponse{Event: event}, nil
}
