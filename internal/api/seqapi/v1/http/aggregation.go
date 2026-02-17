package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveGetAggregation go doc.
//
//	@Router		/seqapi/v1/aggregation [post]
//	@ID			seqapi_v1_getAggregation
//	@Tags		seqapi_v1
//	@Accept		json
//	@Produce	json
//	@Param		env		query		string					false	"Environment"
//	@Param		body	body		getAggregationRequest	true	"Request body"
//	@Success	200		{object}	getAggregationResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetAggregation(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_aggregation")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getAggregationRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse getAggregation request: %w", err), http.StatusBadRequest)
		return
	}

	aggsRaw, _ := json.Marshal(httpReq.Aggregations)
	env := getEnvFromContext(ctx)
	client, options, err := a.GetClientFromEnv(env)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}
	span.SetAttributes(
		attribute.KeyValue{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		attribute.KeyValue{
			Key:   "from",
			Value: attribute.StringValue(httpReq.From.Format(time.DateTime)),
		},
		attribute.KeyValue{
			Key:   "to",
			Value: attribute.StringValue(httpReq.To.Format(time.DateTime)),
		},
		attribute.KeyValue{
			Key:   "agg_field",
			Value: attribute.StringValue(httpReq.AggField),
		},
		attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggsRaw)),
		},
		attribute.KeyValue{
			Key:   "env",
			Value: attribute.StringValue(env),
		},
	)

	if err := api_error.CheckAggregationsCount(len(httpReq.Aggregations), options.MaxAggregationsPerRequest); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}
	md := metadata.New(map[string]string{
		"env": env,
	})
	grpcCtx := metadata.NewOutgoingContext(ctx, md)
	resp, err := client.GetAggregation(grpcCtx, httpReq.toProto())
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	getAggResp := getAggregationResponseFromProto(resp)

	if a.masker != nil {
		buf := make([]string, 0)
		for i, agg := range getAggResp.Aggregations {
			buf = buf[:0]
			for _, b := range agg.Buckets {
				buf = append(buf, b.Key)
			}

			aggReq := httpReq.Aggregations[i]
			field := aggReq.Field
			if aggReq.GroupBy != "" {
				field = aggReq.GroupBy
			}

			buf = a.masker.MaskAgg(field, buf)

			for j, key := range buf {
				getAggResp.Aggregations[i].Buckets[j].Key = key
			}
		}
	}

	wr.WriteJson(getAggResp)
}

type aggregationFunc string //	@name	seqapi.v1.AggregationFunc

const (
	afCount    aggregationFunc = "count"
	afSum      aggregationFunc = "sum"
	afMin      aggregationFunc = "min"
	afMax      aggregationFunc = "max"
	afAvg      aggregationFunc = "avg"
	afQuantile aggregationFunc = "quantile"
	afUnique   aggregationFunc = "unique"
)

func (f aggregationFunc) toProto() seqapi.AggFunc {
	switch f {
	case afSum:
		return seqapi.AggFunc_AGG_FUNC_SUM
	case afMin:
		return seqapi.AggFunc_AGG_FUNC_MIN
	case afMax:
		return seqapi.AggFunc_AGG_FUNC_MAX
	case afAvg:
		return seqapi.AggFunc_AGG_FUNC_AVG
	case afQuantile:
		return seqapi.AggFunc_AGG_FUNC_QUANTILE
	case afUnique:
		return seqapi.AggFunc_AGG_FUNC_UNIQUE
	default:
		return seqapi.AggFunc_AGG_FUNC_COUNT
	}
}

func aggregationFuncFromProto(f seqapi.AggFunc) aggregationFunc {
	switch f {
	case seqapi.AggFunc_AGG_FUNC_COUNT:
		return afCount
	case seqapi.AggFunc_AGG_FUNC_SUM:
		return afSum
	case seqapi.AggFunc_AGG_FUNC_MIN:
		return afMin
	case seqapi.AggFunc_AGG_FUNC_MAX:
		return afMax
	case seqapi.AggFunc_AGG_FUNC_AVG:
		return afAvg
	case seqapi.AggFunc_AGG_FUNC_QUANTILE:
		return afQuantile
	case seqapi.AggFunc_AGG_FUNC_UNIQUE:
		return afUnique
	default:
		return afCount
	}
}

type aggregationQuery struct {
	Field     string          `json:"field"`
	GroupBy   string          `json:"group_by"`
	Func      aggregationFunc `json:"agg_func" default:"count"`
	Quantiles []float64       `json:"quantiles,omitempty"`
} //	@name	seqapi.v1.AggregationQuery

func (aq aggregationQuery) toProto() *seqapi.AggregationQuery {
	return &seqapi.AggregationQuery{
		Field:     aq.Field,
		GroupBy:   aq.GroupBy,
		Func:      aq.Func.toProto(),
		Quantiles: aq.Quantiles,
	}
}

type aggregationQueries []aggregationQuery

func (aqs aggregationQueries) toProto() []*seqapi.AggregationQuery {
	if len(aqs) == 0 {
		return nil
	}
	res := make([]*seqapi.AggregationQuery, len(aqs))
	for i, aq := range aqs {
		res[i] = aq.toProto()
	}
	return res
}

type getAggregationRequest struct {
	Query        string             `json:"query"`
	From         time.Time          `json:"from" format:"date-time"`
	To           time.Time          `json:"to" format:"date-time"`
	AggField     string             `json:"aggField"`
	Aggregations aggregationQueries `json:"aggregations"`
} //	@name	seqapi.v1.GetAggregationRequest

func (r getAggregationRequest) toProto() *seqapi.GetAggregationRequest {
	return &seqapi.GetAggregationRequest{
		Query:        r.Query,
		From:         timestamppb.New(r.From),
		To:           timestamppb.New(r.To),
		AggField:     r.AggField,
		Aggregations: r.Aggregations.toProto(),
	}
}

type aggregationBucket struct {
	Key       string    `json:"key"`
	Value     *float64  `json:"value"`
	NotExists int64     `json:"not_exists,omitempty"`
	Quantiles []float64 `json:"quantiles,omitempty"`
} //	@name	seqapi.v1.AggregationBucket

func aggregationBucketFromProto(proto *seqapi.Aggregation_Bucket) aggregationBucket {
	return aggregationBucket{
		Key:       proto.GetKey(),
		Value:     proto.Value,
		NotExists: proto.GetNotExists(),
		Quantiles: proto.GetQuantiles(),
	}
}

type aggregationBuckets []aggregationBucket

func aggregationBucketsFromProto(proto []*seqapi.Aggregation_Bucket) aggregationBuckets {
	res := make(aggregationBuckets, len(proto))
	for i, ab := range proto {
		res[i] = aggregationBucketFromProto(ab)
	}
	return res
}

type aggregation struct {
	Buckets   aggregationBuckets `json:"buckets"`
	NotExists int64              `json:"not_exists,omitempty"`
} //	@name	seqapi.v1.Aggregation

func aggregationFromProto(proto *seqapi.Aggregation) aggregation {
	return aggregation{
		Buckets:   aggregationBucketsFromProto(proto.GetBuckets()),
		NotExists: proto.GetNotExists(),
	}
}

type aggregations []aggregation

func aggregationsFromProto(proto []*seqapi.Aggregation, alwaysCreate bool) aggregations {
	if len(proto) == 0 && !alwaysCreate {
		return nil
	}
	res := make(aggregations, len(proto))
	for i, a := range proto {
		res[i] = aggregationFromProto(a)
	}
	return res
}

type getAggregationResponse struct {
	Aggregation     aggregation  `json:"aggregation"`
	Aggregations    aggregations `json:"aggregations"`
	Error           apiError     `json:"error"`
	PartialResponse bool         `json:"partialResponse"`
} //	@name	seqapi.v1.GetAggregationResponse

func getAggregationResponseFromProto(proto *seqapi.GetAggregationResponse) getAggregationResponse {
	return getAggregationResponse{
		Aggregation:     aggregationFromProto(proto.GetAggregation()),
		Aggregations:    aggregationsFromProto(proto.GetAggregations(), true),
		Error:           apiErrorFromProto(proto.GetError()),
		PartialResponse: proto.GetPartialResponse(),
	}
}
