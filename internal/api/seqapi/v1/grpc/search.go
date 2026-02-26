package grpc

import (
	"context"
	"encoding/json"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) Search(ctx context.Context, req *seqapi.SearchRequest) (*seqapi.SearchResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_search")
	defer span.End()

	env := a.GetEnvFromContext(ctx)

	spanAttributes := []attribute.KeyValue{
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
			Key:   "with_total",
			Value: attribute.BoolValue(req.GetWithTotal()),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(req.GetOffset())),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(req.GetLimit())),
		},
		{
			Key:   "order",
			Value: attribute.StringValue(req.GetOrder().String()),
		},
	}

	if env != "" {
		spanAttributes = append(spanAttributes, attribute.String("env", env))
	}

	span.SetAttributes(spanAttributes...)

	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Histogram != nil && req.Histogram.Interval != "" {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "histogram_interval",
			Value: attribute.StringValue(req.Histogram.Interval),
		})
	}
	if len(req.Aggregations) > 0 {
		aggregations, _ := json.Marshal(req.Aggregations)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggregations)),
		})
	}

	span.SetAttributes(spanAttributes...)

	if err := api_error.CheckSearchLimit(req.Limit, params.options.MaxSearchLimit); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := api_error.CheckAggregationsCount(len(req.Aggregations), params.options.MaxAggregationsPerRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := api_error.CheckSearchOffsetLimit(req.Offset, params.options.MaxSearchOffsetLimit); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fromRaw, toRaw := req.From.AsTime(), req.To.AsTime()
	for _, agg := range req.Aggregations {
		if agg.Interval == nil {
			continue
		}
		if err := api_error.CheckAggregationTsInterval(*agg.Interval, fromRaw, toRaw,
			params.options.MaxBucketsPerAggregationTs,
		); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	resp, err := params.client.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Total > params.options.MaxSearchTotalLimit {
		resp.Error = &seqapi.Error{
			Code:    seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY,
			Message: api_error.ErrQueryTooHeavy.Error(),
		}
	}

	if params.masker != nil {
		for _, e := range resp.Events {
			params.masker.Mask(e.Data)
		}
	}

	return resp, nil
}
