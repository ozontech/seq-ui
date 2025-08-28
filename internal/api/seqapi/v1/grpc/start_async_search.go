package grpc

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) StartAsyncSearch(
	ctx context.Context,
	req *seqapi.StartAsyncSearchRequest,
) (*seqapi.StartAsyncSearchResponse, error) {
	if a.asyncSearches == nil {
		return nil, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error())
	}

	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_start_async_search")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		{
			Key:   "retention",
			Value: attribute.StringValue(req.GetRetention().String()),
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
			Key:   "with_docs",
			Value: attribute.BoolValue(req.GetWithDocs()),
		},
	}
	if req.Hist != nil && req.Hist.Interval != "" {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "histogram_interval",
			Value: attribute.StringValue(req.Hist.Interval),
		})
	}
	if len(req.Aggs) > 0 {
		aggregations, _ := json.Marshal(req.Aggs)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggregations)),
		})
	}

	span.SetAttributes(spanAttributes...)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	resp, err := a.asyncSearches.StartAsyncSearch(ctx, profileID, req)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return resp, nil
}
