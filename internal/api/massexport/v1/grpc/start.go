package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) Start(ctx context.Context, req *massexport_v1.StartRequest) (*massexport_v1.StartResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_start")
	defer span.End()

	env := types.GetEnvKey(ctx)
	svc := a.serviceByEnv[env]
	from := req.GetFrom().AsTime()
	to := req.GetTo().AsTime()

	attributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		{
			Key:   "from",
			Value: attribute.StringValue(from.Format(time.RFC3339)),
		},
		{
			Key:   "to",
			Value: attribute.StringValue(to.Format(time.RFC3339)),
		},
		{
			Key:   "name",
			Value: attribute.StringValue(req.GetName()),
		},
	}
	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}
	span.SetAttributes(attributes...)

	if req.GetName() == "" {
		return nil, grpcutil.ProcessError(types.NewErrInvalidRequestField("empty export name"))
	}

	if !(from.Before(to)) {
		return nil, grpcutil.ProcessError(types.NewErrInvalidRequestField("'from' is not before 'to'"))
	}

	result, err := svc.StartExport(ctx, types.StartExportRequest{
		Query: req.GetQuery(),
		From:  from,
		To:    to,
		Name:  req.GetName(),
	})
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &massexport_v1.StartResponse{
		SessionId: result.SessionID,
	}, nil
}
