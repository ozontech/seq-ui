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
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_releases")
	defer span.End()

	svc, reqEnv := a.serviceByEnv(ctx, req.Env)

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
	}
	if reqEnv != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*reqEnv)})
	}
	span.SetAttributes(attributes...)

	request := types.GetReleasesRequest{
		Service: req.Service,
		Env:     reqEnv,
	}
	releases, err := svc.GetReleases(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetReleasesResponse{Releases: releases}, nil
}
