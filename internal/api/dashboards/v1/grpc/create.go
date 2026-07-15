package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
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

	request := types.CreateDashboardRequest{
		Name: req.Name,
		Meta: req.Meta,
	}

	uuid, err := a.service.CreateDashboard(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &dashboards.CreateResponse{
		Uuid: uuid,
	}, nil
}
