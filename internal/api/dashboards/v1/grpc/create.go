package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Create(ctx context.Context, req *dashboards.CreateRequest) (*dashboards.CreateResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_create")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(req.GetName()),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.CreateDashboardRequest{
		ProfileID: profileID,
		Name:      req.Name,
		Meta:      req.Meta,
	}
	uuid, err := a.service.CreateDashboard(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &dashboards.CreateResponse{
		Uuid: uuid,
	}, nil
}
