package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Cancel(ctx context.Context, req *massexport.CancelRequest) (*massexport.CancelResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_cancel")
	defer span.End()

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(req.GetSessionId()),
	})

	err := a.exporter.CancelExport(ctx, req.GetSessionId())
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &massexport.CancelResponse{}, nil
}
