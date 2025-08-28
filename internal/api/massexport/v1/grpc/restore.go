package grpc

import (
	"context"
	"fmt"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Restore(ctx context.Context, req *massexport_v1.RestoreRequest) (*massexport_v1.RestoreResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_restore")
	defer span.End()

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(req.GetSessionId()),
	})

	err := a.exporter.RestoreExport(ctx, req.GetSessionId())
	if err != nil {
		return nil, grpcutil.ProcessError(fmt.Errorf("restore export: %w", err))
	}

	return &massexport_v1.RestoreResponse{}, nil
}
