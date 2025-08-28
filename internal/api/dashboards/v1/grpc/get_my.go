package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) GetMy(ctx context.Context, req *dashboards.GetMyRequest) (*dashboards.GetMyResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_get_my")
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

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.GetUserDashboardsRequest{
		ProfileID: profileID,
		Limit:     int(req.Limit),
		Offset:    int(req.Offset),
	}
	dis, err := a.service.GetMyDashboards(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	ds := make([]*dashboards.GetMyResponse_Dashboard, len(dis))
	for i, d := range dis {
		ds[i] = &dashboards.GetMyResponse_Dashboard{
			Uuid: d.UUID,
			Name: d.Name,
		}
	}

	return &dashboards.GetMyResponse{
		Dashboards: ds,
	}, nil
}
