package grpc

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) Restore(ctx context.Context, req *massexport_v1.RestoreRequest) (*massexport_v1.RestoreResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_restore")
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

	err := svc.RestoreExport(ctx, req.GetSessionId())
	if err != nil {
		return nil, grpcutil.ProcessError(fmt.Errorf("restore export: %w", err))
	}

	return &massexport_v1.RestoreResponse{}, nil
}
