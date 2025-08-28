package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) GetAll(ctx context.Context, req *dashboards.GetAllRequest) (*dashboards.GetAllResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_get_all")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(int(req.GetLimit())),
		},
		attribute.KeyValue{
			Key:   "offset",
			Value: attribute.IntValue(int(req.GetOffset())),
		},
	)

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.GetAllDashboardsRequest{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	dis, err := a.service.GetAllDashboards(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	ds := make([]*dashboards.GetAllResponse_Dashboard, len(dis))
	for i, d := range dis {
		ds[i] = &dashboards.GetAllResponse_Dashboard{
			Uuid:      d.UUID,
			Name:      d.Name,
			OwnerName: d.OwnerName,
		}
	}

	return &dashboards.GetAllResponse{
		Dashboards: ds,
	}, nil
}
