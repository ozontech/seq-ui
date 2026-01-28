package grpc

import (
	"context"
	"encoding/json"
	"time"

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

	if err := api_error.CheckSearchLimit(req.Limit, a.config.MaxSearchLimit); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := api_error.CheckAggregationsCount(len(req.Aggregations), a.config.MaxAggregationsPerRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := api_error.CheckSearchOffsetLimit(req.Offset, a.config.MaxSearchOffsetLimit); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	aggIntervals := make([]*string, 0, len(req.Aggregations))
	fromRaw, toRaw := req.From.AsTime(), req.To.AsTime()
	for _, agg := range req.Aggregations {
		if agg.Interval == nil {
			aggIntervals = append(aggIntervals, nil)
			continue
		}
		if err := api_error.CheckAggregationTsInterval(*agg.Interval, fromRaw, toRaw,
			a.config.MaxBucketsPerAggregationTs,
		); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if agg.Func != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggIntervals = append(aggIntervals, nil)
			continue
		}

		aggIntervals = append(aggIntervals, agg.Interval)
	}

	resp, err := a.seqDB.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Total > a.config.MaxSearchTotalLimit {
		resp.Error = &seqapi.Error{
			Code:    seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY,
			Message: api_error.ErrQueryTooHeavy.Error(),
		}
	}

	if a.masker != nil {
		for _, e := range resp.Events {
			a.masker.Mask(e.Data)
		}
	}

	for i, agg := range resp.Aggregations {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == nil {
			continue
		}

		interval, err := time.ParseDuration(*aggIntervals[i])
		if err != nil {
			return nil, err
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}

			*bucket.Value /= interval.Seconds()
		}
	}

	return resp, nil
}
