package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetReleases(ctx context.Context, req *errorgroups.GetReleasesRequest) (*errorgroups.GetReleasesResponse, error) {
	_, span := tracing.StartSpan(ctx, "errorgroups_v1_get_groups")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
	}
	if req.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*req.Env)})
	}
	span.SetAttributes(attributes...)

	request := types.GetErrorGroupReleasesRequest{
		Service: req.Service,
		Env:     req.Env,
	}
	releases, err := a.service.GetReleases(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetReleasesResponse{Releases: releases}, nil
}
