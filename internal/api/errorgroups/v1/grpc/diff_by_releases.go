package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *API) DiffByReleases(ctx context.Context, req *errorgroups.DiffByReleasesRequest) (*errorgroups.DiffByReleasesResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_diff_by_releases")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
		{Key: "releases", Value: attribute.StringSliceValue(req.Releases)},
		{Key: "limit", Value: attribute.IntValue(int(req.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(req.Offset))},
		{Key: "order", Value: attribute.StringValue(req.Order.String())},
		{Key: "with_total", Value: attribute.BoolValue(req.WithTotal)},
	}
	if req.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*req.Env)})
	}
	if req.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*req.Source)})
	}
	span.SetAttributes(attributes...)

	groups, total, err := a.service.DiffByReleases(ctx, types.DiffByReleasesRequest{
		Service:   req.Service,
		Releases:  req.Releases,
		Env:       req.Env,
		Source:    req.Source,
		Limit:     req.Limit,
		Offset:    req.Offset,
		Order:     types.ErrorGroupsOrder(req.Order),
		WithTotal: req.WithTotal,
	})
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.DiffByReleasesResponse{
		Total:  total,
		Groups: diffGroupsToProto(groups),
	}, nil
}

func diffGroupsToProto(source []types.DiffGroup) []*errorgroups.DiffByReleasesResponse_Group {
	groups := make([]*errorgroups.DiffByReleasesResponse_Group, 0, len(source))

	for _, g := range source {
		releaseInfos := make(map[string]*errorgroups.DiffByReleasesResponse_ReleaseInfo)
		for release, info := range g.ReleaseInfos {
			releaseInfos[release] = &errorgroups.DiffByReleasesResponse_ReleaseInfo{
				SeenTotal: info.SeenTotal,
			}
		}
		groups = append(groups, &errorgroups.DiffByReleasesResponse_Group{
			Hash:         g.Hash,
			Message:      g.Message,
			FirstSeenAt:  timestamppb.New(g.FirstSeenAt),
			LastSeenAt:   timestamppb.New(g.LastSeenAt),
			Source:       g.Source,
			ReleaseInfos: releaseInfos,
		})
	}

	return groups
}
