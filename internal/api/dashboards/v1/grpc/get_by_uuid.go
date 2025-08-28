package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) GetByUUID(ctx context.Context, req *dashboards.GetByUUIDRequest) (*dashboards.GetByUUIDResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_get_by_uuid")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "uuid",
			Value: attribute.StringValue(req.GetUuid()),
		},
	)

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	d, err := a.service.GetDashboardByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &dashboards.GetByUUIDResponse{
		Name:      d.Name,
		Meta:      d.Meta,
		OwnerName: d.OwnerName,
	}, nil
}
