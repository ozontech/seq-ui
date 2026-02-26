package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetHistogram(ctx context.Context, req *seqapi.GetHistogramRequest) (*seqapi.GetHistogramResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_histogram")
	defer span.End()

	env := a.GetEnvFromContext(ctx)

	attributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		{
			Key:   "from",
			Value: tracing.TimestampToStringValue(req.GetFrom()),
		},
		{
			Key:   "to",
			Value: tracing.TimestampToStringValue(req.GetTo()),
		},
		{
			Key:   "interval",
			Value: attribute.StringValue(req.GetInterval()),
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

	resp, err := params.client.GetHistogram(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
