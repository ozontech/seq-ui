package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetGroups(ctx context.Context, req *errorgroups.GetGroupsRequest) (*errorgroups.GetGroupsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_groups")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
		{Key: "limit", Value: attribute.IntValue(int(req.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(req.Offset))},
		{Key: "order", Value: attribute.StringValue(string(req.Order))},
	}
	if req.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*req.Env)})
	}
	if req.Release != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "release", Value: attribute.StringValue(*req.Release)})
	}
	if req.Duration != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "duration", Value: attribute.StringValue(req.Duration.String())})
	}
	if req.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*req.Source)})
	}
	span.SetAttributes(attributes...)

	var duration *time.Duration
	if req.Duration != nil {
		parsedDuration := req.Duration.AsDuration()
		duration = &parsedDuration
	}

	request := types.GetErrorGroupsRequest{
		Service:   req.Service,
		Env:       req.Env,
		Source:    req.Source,
		Release:   req.Release,
		Duration:  duration,
		Limit:     req.Limit,
		Offset:    req.Offset,
		Order:     types.ErrorGroupsOrder(req.Order),
		WithTotal: req.WithTotal,
	}
	groups, total, err := a.service.GetErrorGroups(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetGroupsResponse{
		Total:  total,
		Groups: groupsToProto(groups),
	}, nil
}

func groupsToProto(source []types.ErrorGroup) []*errorgroups.Group {
	groups := make([]*errorgroups.Group, 0, len(source))

	for _, g := range source {
		groups = append(groups, &errorgroups.Group{
			Hash:        g.Hash,
			Message:     g.Message,
			SeenTotal:   g.SeenTotal,
			FirstSeenAt: timestamppb.New(g.FirstSeenAt),
			LastSeenAt:  timestamppb.New(g.LastSeenAt),
			Source:      g.Source,
		})
	}

	return groups
}
