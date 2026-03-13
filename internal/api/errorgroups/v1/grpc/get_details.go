package grpc

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetDetails(ctx context.Context, req *errorgroups.GetDetailsRequest) (*errorgroups.GetDetailsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_details")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
		{Key: "group_hash", Value: attribute.StringValue(strconv.FormatUint(req.GroupHash, 10))},
	}
	if req.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*req.Env)})
	}
	if req.Release != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "release", Value: attribute.StringValue(*req.Release)})
	}
	if req.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*req.Source)})
	}
	span.SetAttributes(attributes...)

	request := types.GetErrorGroupDetailsRequest{
		Service:   req.Service,
		GroupHash: req.GroupHash,
		Env:       req.Env,
		Source:    req.Source,
		Release:   req.Release,
	}
	details, err := a.service.GetDetails(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetDetailsResponse{
		GroupHash:     details.GroupHash,
		Message:       details.Message,
		SeenTotal:     details.SeenTotal,
		FirstSeenAt:   timestamppb.New(details.FirstSeenAt),
		LastSeenAt:    timestamppb.New(details.LastSeenAt),
		LogTags:       details.LogTags,
		Distributions: distributionsToProto(details.Distributions),
		Source:        details.Source,
	}, nil
}

func distributionsToProto(source types.ErrorGroupDistributions) *errorgroups.GetDetailsResponse_Distributions {
	distrToProto := func(d types.ErrorGroupDistribution) *errorgroups.GetDetailsResponse_Distribution {
		return &errorgroups.GetDetailsResponse_Distribution{
			Value:   d.Value,
			Percent: d.Percent,
		}
	}

	ds := &errorgroups.GetDetailsResponse_Distributions{
		ByEnv:     make([]*errorgroups.GetDetailsResponse_Distribution, 0, len(source.ByEnv)),
		ByRelease: make([]*errorgroups.GetDetailsResponse_Distribution, 0, len(source.ByRelease)),
	}

	for _, d := range source.ByEnv {
		ds.ByEnv = append(ds.ByEnv, distrToProto(d))
	}
	for _, d := range source.ByRelease {
		ds.ByRelease = append(ds.ByRelease, distrToProto(d))
	}

	return ds
}
