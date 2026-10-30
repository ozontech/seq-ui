package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetEvent(ctx context.Context, req *seqapi.GetEventRequest) (*seqapi.GetEventResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_event")
	defer span.End()

	env := types.GetEnvKey(ctx)

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

	params := a.paramsByEnv[env]

	if cached, ok := params.caches.Events.Get(ctx, req.Id); ok {
		event := &seqapi.Event{}
		err := proto.Unmarshal([]byte(cached), event)
		if err == nil {
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

	if params.caches.Events != nil {
		if data, err := proto.Marshal(resp.Event); err == nil {
			params.caches.Events.Set(ctx, req.Id, string(data))
		} else {
			logger.Error("failed to marshal event proto for caching", zap.String("id", req.Id), zap.Error(err))
		}
	}

	return resp, nil
}
