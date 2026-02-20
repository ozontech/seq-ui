package grpc

import (
	"context"
	"encoding/json"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/aggregation_ts"
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

	aggIntervals, err := aggregation_ts.GetIntervals(req.Aggregations)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	bucketUnits, err := aggregation_ts.GetBucketUnits(req.Aggregations, a.config.DefaultBucketUnit)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	aggregation_ts.NormalizeBucketValues(resp.Aggregations, aggIntervals, bucketUnits)

	return resp, nil
}
