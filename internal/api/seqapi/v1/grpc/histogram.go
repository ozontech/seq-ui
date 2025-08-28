package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func (a *API) GetHistogram(ctx context.Context, req *seqapi.GetHistogramRequest) (*seqapi.GetHistogramResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_histogram")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		attribute.KeyValue{
			Key:   "from",
			Value: tracing.TimestampToStringValue(req.GetFrom()),
		},
		attribute.KeyValue{
			Key:   "to",
			Value: tracing.TimestampToStringValue(req.GetTo()),
		},
		attribute.KeyValue{
			Key:   "interval",
			Value: attribute.StringValue(req.GetInterval()),
		},
	)

	resp, err := a.seqDB.GetHistogram(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
