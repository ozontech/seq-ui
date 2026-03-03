package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (a *API) GetEvent(ctx context.Context, req *seqapi.GetEventRequest) (*seqapi.GetEventResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_event")
	defer span.End()

	env := a.GetEnvFromContext(ctx)

	attributes := []attribute.KeyValue{
		{
			Key:   "id",
			Value: attribute.StringValue(req.GetId()),
		},
	}

	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}

	span.SetAttributes(attributes...)

	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if cached, err := a.inmemWithRedisCache.Get(ctx, req.Id); err == nil {
		event := &seqapi.Event{}
		if err = proto.Unmarshal([]byte(cached), event); err == nil {
			if params.masker != nil {
				params.masker.Mask(event.Data)
			}
			return &seqapi.GetEventResponse{Event: event}, nil
		}
		logger.Error("failed to unmarshal cached event proto", zap.String("id", req.Id), zap.Error(err))
	}
	resp, err := params.client.GetEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	if params.masker != nil && resp.Event != nil {
		params.masker.Mask(resp.Event.Data)
	}

	if data, err := proto.Marshal(resp.Event); err == nil {
		_ = a.inmemWithRedisCache.SetWithTTL(ctx, req.Id, string(data), params.options.EventsCacheTTL)
	} else {
		logger.Error("failed to marshal event proto for caching", zap.String("id", req.Id), zap.Error(err))
	}

	return resp, nil
}
