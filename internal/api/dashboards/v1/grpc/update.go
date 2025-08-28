package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Update(ctx context.Context, req *dashboards.UpdateRequest) (*dashboards.UpdateResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_update")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "uuid",
			Value: attribute.StringValue(req.GetUuid()),
		},
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(req.GetName()),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.UpdateDashboardRequest{
		UUID:      req.Uuid,
		ProfileID: profileID,
		Name:      req.Name,
		Meta:      req.Meta,
	}
	if err = a.service.UpdateDashboard(ctx, request); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &dashboards.UpdateResponse{}, nil
}
