package grpc

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/errorgroups/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetHist(ctx context.Context, req *errorgroups.GetHistRequest) (*errorgroups.GetHistResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "errorgroups_v1_get_groups")
	defer span.End()

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(req.Service)},
	}
	if req.GroupHash != nil {
		attributes = append(attributes, attribute.KeyValue{
			Key: "group_hash", Value: attribute.StringValue(strconv.FormatUint(*req.GroupHash, 10)),
		})
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

	request := types.GetErrorHistRequest{
		Service:   req.Service,
		GroupHash: req.GroupHash,
		Env:       req.Env,
		Source:    req.Source,
		Release:   req.Release,
		Duration:  duration,
	}
	buckets, err := a.service.GetHist(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &errorgroups.GetHistResponse{
		Buckets: bucketsToProto(buckets),
	}, nil
}

func bucketsToProto(source []types.ErrorHistBucket) []*errorgroups.Bucket {
	buckets := make([]*errorgroups.Bucket, 0, len(source))

	for _, b := range source {
		buckets = append(buckets, &errorgroups.Bucket{
			Time:  timestamppb.New(b.Time),
			Count: b.Count,
		})
	}

	return buckets
}
