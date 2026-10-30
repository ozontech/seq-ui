package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetServiceFilters(ctx context.Context, req *errorgroups.GetServiceFiltersRequest) (*errorgroups.GetServiceFiltersResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_service_filters")
	defer span.End()

	svc, reqEnv := a.serviceByEnv(ctx, req.Env)

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
	}
	if reqEnv != nil && *reqEnv != "" {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*reqEnv)})
	}
	span.SetAttributes(attributes...)

	request := types.GetServiceFiltersRequest{
		Service: req.Service,
		Env:     reqEnv,
	}
	filters, err := svc.GetServiceFilters(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetServiceFiltersResponse{
		Filters: filtersToProto(filters),
	}, nil
}

func filtersToProto(source []types.ServiceFilter) []*errorgroups.GetServiceFiltersResponse_Filter {
	filters := make([]*errorgroups.GetServiceFiltersResponse_Filter, 0, len(source))

	for _, f := range source {
		filters = append(filters, &errorgroups.GetServiceFiltersResponse_Filter{
			Key:    f.Key,
			Values: f.Values,
		})
	}

	return filters
}
