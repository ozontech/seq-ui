package grpc

import (
	"context"
	"encoding/json"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) Search(ctx context.Context, req *dashboards.SearchRequest) (*dashboards.SearchResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "dashboards_v1_search")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(req.GetLimit())),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(req.GetOffset())),
		},
	}
	if req.GetFilter() != nil {
		filter, _ := json.Marshal(req.Filter)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "filter",
			Value: attribute.StringValue(string(filter)),
		})
	}

	span.SetAttributes(spanAttributes...)

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.SearchDashboardsRequest{
		Query:  req.Query,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	if req.Filter != nil {
		request.Filter = &types.SearchDashboardsFilter{
			OwnerName: req.Filter.OwnerName,
		}
	}

	infos, err := a.service.SearchDashboards(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	ds := make([]*dashboards.SearchResponse_Dashboard, len(infos))
	for i, d := range infos {
		ds[i] = &dashboards.SearchResponse_Dashboard{
			Uuid:      d.UUID,
			Name:      d.Name,
			OwnerName: d.OwnerName,
		}
	}

	return &dashboards.SearchResponse{
		Dashboards: ds,
	}, nil
}
