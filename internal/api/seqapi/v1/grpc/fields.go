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
)

func (a *API) GetFields(ctx context.Context, req *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_fields")
	defer span.End()

	env, err := a.GetEnvFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	client, _, err := a.GetClientFromEnv(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	span.SetAttributes(attribute.KeyValue{
		Key:   "env",
		Value: attribute.StringValue(env),
	})

	if a.fieldsCache == nil {
		return client.GetFields(ctx, req)
	}

	fields, cached, isActual := a.fieldsCache.getFields()
	if cached && isActual {
		return &seqapi.GetFieldsResponse{
			Fields: fields,
		}, nil
	}

	resp, err := client.GetFields(ctx, req)
	if err != nil {
		if cached {
			logger.Error("can't get fields; use cached fields", zap.Error(err))
			return &seqapi.GetFieldsResponse{Fields: fields}, nil
		}

		return nil, err
	}

	a.fieldsCache.setFields(resp.GetFields())
	return resp, nil
}

func (a *API) GetPinnedFields(_ context.Context, _ *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	return &seqapi.GetFieldsResponse{
		Fields: a.pinnedFields,
	}, nil
}
