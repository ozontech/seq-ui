package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Delete(ctx context.Context, req *dashboards.DeleteRequest) (*dashboards.DeleteResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_delete")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "uuid",
			Value: attribute.StringValue(req.GetUuid()),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.DeleteDashboardRequest{
		UUID:      req.Uuid,
		ProfileID: profileID,
	}
	if err = a.service.DeleteDashboard(ctx, request); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &dashboards.DeleteResponse{}, nil
}
