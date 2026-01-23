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

func (a *API) GetAggregation(ctx context.Context, req *seqapi.GetAggregationRequest) (*seqapi.GetAggregationResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_aggregation")
	defer span.End()

	aggregations, _ := json.Marshal(req.Aggregations)

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
			Key:   "agg_field",
			Value: attribute.StringValue(req.GetAggField()),
		},
		attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggregations)),
		},
	)

	if err := api_error.CheckAggregationsCount(len(req.Aggregations), a.config.MaxAggregationsPerRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	aggIntervals := make([]*string, 0, len(req.Aggregations))
	fromRaw, toRaw := req.From.AsTime(), req.To.AsTime()
	for _, agg := range req.Aggregations {
		if agg.Interval == nil {
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

	resp, err := a.seqDB.GetAggregation(ctx, req)
	if err != nil {
		return nil, err
	}

	if a.masker != nil {
		buf := make([]string, 0)
		for i, agg := range resp.Aggregations {
			if agg == nil {
				continue
			}

			buf = buf[:0]
			for _, b := range agg.Buckets {
				buf = append(buf, b.GetKey())
			}

			aggReq := req.Aggregations[i]
			field := aggReq.Field
			if aggReq.GroupBy != "" {
				field = aggReq.GroupBy
			}

			buf = a.masker.MaskAgg(field, buf)

			for j, key := range buf {
				if agg.Buckets[j] != nil {
					agg.Buckets[j].Key = key
				}
			}
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
