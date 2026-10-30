package grpc

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetTopGroups(ctx context.Context, req *errorgroups.GetTopGroupsRequest) (*errorgroups.GetTopGroupsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_top_groups")
	defer span.End()

	svc, reqEnv := a.serviceByEnv(ctx, req.Env)

	attributes := []attribute.KeyValue{
		{Key: "limit", Value: attribute.IntValue(int(req.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(req.Offset))},
		{Key: "with_total", Value: attribute.BoolValue(req.WithTotal)},
	}
	if reqEnv != nil && *reqEnv != "" {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*reqEnv)})
	}
	if req.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*req.Source)})
	}
	if req.Duration != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "duration", Value: attribute.StringValue(req.Duration.String())})
	}
	if req.TimeRange != nil {
		trRaw, _ := json.Marshal(req.TimeRange)
		attributes = append(attributes, attribute.KeyValue{Key: "time_range", Value: attribute.StringValue(string(trRaw))})
	}
	span.SetAttributes(attributes...)

	request := types.GetTopErrorGroupsRequest{
		Env:       reqEnv,
		Source:    req.Source,
		TimeRange: parseTimeRange(req),
		Limit:     req.Limit,
		Offset:    req.Offset,
		WithTotal: req.WithTotal,
	}

	groups, total, err := svc.GetTopErrorGroups(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetTopGroupsResponse{
		Total:  total,
		Groups: topGroupsToProto(groups),
	}, nil
}

func topGroupsToProto(source []types.TopErrorGroup) []*errorgroups.GetTopGroupsResponse_Group {
	groups := make([]*errorgroups.GetTopGroupsResponse_Group, 0, len(source))

	for _, g := range source {
		groups = append(groups, &errorgroups.GetTopGroupsResponse_Group{
			Hash:      g.Hash,
			Message:   g.Message,
			Source:    g.Source,
			SeenTotal: g.Count,
		})
	}

	return groups
}
