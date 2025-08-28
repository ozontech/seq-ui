package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveGetAggregationTs go doc.
//
//	@Router		/seqapi/v1/aggregation_ts [post]
//	@ID			seqapi_v1_getAggregationTs
//	@Tags		seqapi_v1
//	@Accept		json
//	@Produce	json
//	@Param		body	body		getAggregationTsRequest		true	"Request body"
//	@Success	200		{object}	getAggregationTsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetAggregationTs(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_aggregation_ts")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getAggregationTsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse getAggregationTs request: %w", err), http.StatusBadRequest)
		return
	}

	aggsRaw, _ := json.Marshal(httpReq.Aggregations)

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
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggsRaw)),
		},
	)

	if err := api_error.CheckAggregationsCount(len(httpReq.Aggregations), a.config.MaxAggregationsPerRequest); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}
	for _, agg := range httpReq.Aggregations {
		if err := api_error.CheckAggregationTsInterval(agg.Interval, httpReq.From, httpReq.To,
			a.config.MaxBucketsPerAggregationTs,
		); err != nil {
			wr.Error(err, http.StatusBadRequest)
			return
		}
	}

	resp, err := a.seqDB.GetAggregation(ctx, httpReq.toProto())
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(getAggregationTsResponseFromProto(resp, httpReq.Aggregations))
}

type aggregationTsQuery struct {
	aggregationQuery
	Interval string `json:"interval" format:"duration" example:"1m"`
} //	@name	seqapi.v1.AggregationTsQuery

func (aq aggregationTsQuery) toProto() *seqapi.AggregationQuery {
	q := aq.aggregationQuery.toProto()

	q.Interval = new(string)
	*q.Interval = aq.Interval

	return q
}

type aggregationTsQueries []aggregationTsQuery

func (aqs aggregationTsQueries) toProto() []*seqapi.AggregationQuery {
	if len(aqs) == 0 {
		return nil
	}
	res := make([]*seqapi.AggregationQuery, len(aqs))
	for i, aq := range aqs {
		res[i] = aq.toProto()
	}
	return res
}

type getAggregationTsRequest struct {
	Query        string               `json:"query"`
	From         time.Time            `json:"from" format:"date-time"`
	To           time.Time            `json:"to" format:"date-time"`
	Aggregations aggregationTsQueries `json:"aggregations"`
} //	@name	seqapi.v1.GetAggregationTsRequest

func (r getAggregationTsRequest) toProto() *seqapi.GetAggregationRequest {
	return &seqapi.GetAggregationRequest{
		Query:        r.Query,
		From:         timestamppb.New(r.From),
		To:           timestamppb.New(r.To),
		Aggregations: r.Aggregations.toProto(),
	}
}

type aggregationTsBucket struct {
	Timestamp int64    `json:"timestamp"`
	Value     *float64 `json:"value"`
} //	@name	seqapi.v1.AggregationTsBucket

type aggregationSeries struct {
	Labels  map[string]string     `json:"metric"`
	Buckets []aggregationTsBucket `json:"values"`
} //	@name	seqapi.v1.AggregationSeries

type aggregationsSeries []aggregationSeries

func aggregationsSeriesFromProto(proto []*seqapi.Aggregation_Bucket, reqAgg aggregationTsQuery) aggregationsSeries {
	if len(proto) == 0 {
		return nil
	}

	res := make(aggregationsSeries, 0)
	keyToIdx := make(map[string]int)

	addBucket := func(labels map[string]string, val *float64, ts int64) {
		mapKey := ""
		for _, v := range labels {
			mapKey += v
		}

		idx, ok := keyToIdx[mapKey]
		if !ok {
			res = append(res, aggregationSeries{
				Labels: labels,
			})
			idx = len(res) - 1
			keyToIdx[mapKey] = idx
		}

		res[idx].Buckets = append(res[idx].Buckets, aggregationTsBucket{
			Timestamp: ts,
			Value:     val,
		})
	}

	const quantileLabel = "quantile"
	formatQuantile := func(q float64) string {
		return "p" + strconv.Itoa(int(q*100))
	}

	label := reqAgg.Field
	if reqAgg.Func != afCount && reqAgg.GroupBy != "" {
		label = reqAgg.GroupBy
	}

	for _, b := range proto {
		ts := b.Ts.GetSeconds()

		if len(b.Quantiles) == 0 {
			addBucket(map[string]string{
				label: b.Key,
			}, b.Value, ts)
		}

		for i, q := range b.Quantiles {
			addBucket(map[string]string{
				label:         b.Key,
				quantileLabel: formatQuantile(reqAgg.Quantiles[i]),
			}, &q, ts)
		}
	}
	return res
}

type aggregationTs struct {
	Data struct {
		Series aggregationsSeries `json:"result"`
	} `json:"data"`
} //	@name	seqapi.v1.AggregationTs

func aggregationTsFromProto(proto *seqapi.Aggregation, reqAgg aggregationTsQuery) aggregationTs {
	a := aggregationTs{}
	a.Data.Series = aggregationsSeriesFromProto(proto.Buckets, reqAgg)
	return a
}

type aggregationsTs []aggregationTs

func aggregationsTsFromProto(proto []*seqapi.Aggregation, reqAggs aggregationTsQueries) aggregationsTs {
	if len(proto) == 0 {
		return nil
	}
	res := make(aggregationsTs, len(proto))
	for i, a := range proto {
		res[i] = aggregationTsFromProto(a, reqAggs[i])
	}
	return res
}

type getAggregationTsResponse struct {
	Aggregations aggregationsTs `json:"aggregations"`
	Error        string         `json:"error,omitempty"`
} //	@name	seqapi.v1.GetAggregationTsResponse

func getAggregationTsResponseFromProto(proto *seqapi.GetAggregationResponse, reqAggs aggregationTsQueries) getAggregationTsResponse {
	if proto == nil {
		return getAggregationTsResponse{}
	}
	r := getAggregationTsResponse{
		Aggregations: aggregationsTsFromProto(proto.Aggregations, reqAggs),
	}
	if proto.Error != nil {
		r.Error = proto.Error.Message
	}
	return r
}
