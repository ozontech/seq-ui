package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) Status(ctx context.Context, req *seqapi.StatusRequest) (*seqapi.StatusResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_status")
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

	return client.Status(ctx, req)
}
