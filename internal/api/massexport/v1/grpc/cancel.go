package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) Cancel(ctx context.Context, req *massexport.CancelRequest) (*massexport.CancelResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_cancel")
	defer span.End()

	env := types.GetEnvKey(ctx)
	svc := a.serviceByEnv[env]

	attributes := []attribute.KeyValue{
		{
			Key:   "session_id",
			Value: attribute.StringValue(req.GetSessionId()),
		},
	}
	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}
	span.SetAttributes(attributes...)

	err := svc.CancelExport(ctx, req.GetSessionId())
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &massexport.CancelResponse{}, nil
}
