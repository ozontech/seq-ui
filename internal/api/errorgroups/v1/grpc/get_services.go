package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetServices(ctx context.Context, req *errorgroups.GetServicesRequest) (*errorgroups.GetServicesResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_services")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "query", Value: attribute.StringValue(req.Query)},
	}
	if req.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*req.Env)})
	}
	span.SetAttributes(attributes...)

	request := types.GetServicesRequest{
		Query:  req.Query,
		Env:    req.Env,
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	services, err := a.service.GetServices(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetServicesResponse{Services: services}, nil
}
