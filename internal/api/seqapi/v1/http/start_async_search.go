package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

// serveStartAsyncSearch go doc.
//
//	@Router		/seqapi/v1/async_search/start [post]
//	@ID			seqapi_v1_start_async_search
//	@Tags		seqapi_v1
//	@Param		env		query		string						true	"Environment"
//	@Param		body	body		startAsyncSearchRequest		true	"Request body"
//	@Success	200		{object}	startAsyncSearchResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveStartAsyncSearch(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	if a.asyncSearches == nil {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_start_async_search")
	defer span.End()

	var httpReq startAsyncSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse search request: %w", err), http.StatusBadRequest)
		return
	}

	parsedRetention, err := time.ParseDuration(httpReq.Retention)
	if err != nil {
		wr.Error(fmt.Errorf("failed to parse retention: %w", err), http.StatusBadRequest)
		return
	}

	for _, agg := range httpReq.Aggregations {
		if agg.Interval == "" {
			continue
		}
		if err := api_error.CheckAggregationTsInterval(agg.Interval, httpReq.From, httpReq.To,
			a.config.MaxBucketsPerAggregationTs,
		); err != nil {
			wr.Error(err, http.StatusBadRequest)
			return
		}
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		{
			Key:   "retention",
			Value: attribute.StringValue(httpReq.Retention),
		},
		{
			Key:   "from",
			Value: attribute.StringValue(httpReq.From.Format(time.DateTime)),
		},
		{
			Key:   "to",
			Value: attribute.StringValue(httpReq.To.Format(time.DateTime)),
		},
		{
			Key:   "with_docs",
			Value: attribute.BoolValue(httpReq.WithDocs),
		},
		{
			Key:   "size",
			Value: attribute.Int64Value(int64(httpReq.Size)),
		},
	}
	if httpReq.Histogram != nil && httpReq.Histogram.Interval != "" {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "histogram_interval",
			Value: attribute.StringValue(httpReq.Histogram.Interval),
		})
	}
	if len(httpReq.Aggregations) > 0 {
		aggsRaw, _ := json.Marshal(httpReq.Aggregations)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggsRaw)),
		})
	}

	span.SetAttributes(spanAttributes...)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	resp, err := a.asyncSearches.StartAsyncSearch(ctx, profileID, httpReq.toProto(parsedRetention))
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(startAsyncSearchResponseFromProto(resp))
}

type AsyncSearchRequestHistogram struct {
	Interval string `json:"interval,omitempty"`
}

type startAsyncSearchRequest struct {
	Retention    string                       `json:"retention" format:"duration" example:"1h"`
	Query        string                       `json:"query"`
	From         time.Time                    `json:"from" format:"date-time"`
	To           time.Time                    `json:"to" format:"date-time"`
	Aggregations aggregationTsQueries         `json:"aggregations,omitempty"`
	Histogram    *AsyncSearchRequestHistogram `json:"histogram,omitempty"`
	WithDocs     bool                         `json:"with_docs"`
	Size         int32                        `json:"size"`
	Meta         string                       `json:"meta,omitempty"`
} //	@name	seqapi.v1.StartAsyncSearchRequest

func (r startAsyncSearchRequest) toProto(parsedRetention time.Duration) *seqapi.StartAsyncSearchRequest {
	req := &seqapi.StartAsyncSearchRequest{
		Retention: durationpb.New(parsedRetention),
		Query:     r.Query,
		From:      timestamppb.New(r.From),
		To:        timestamppb.New(r.To),
		Aggs:      r.Aggregations.toProto(),
		WithDocs:  r.WithDocs,
		Size:      r.Size,
		Meta:      r.Meta,
	}
	if r.Histogram != nil && r.Histogram.Interval != "" {
		req.Hist = &seqapi.StartAsyncSearchRequest_HistQuery{
			Interval: r.Histogram.Interval,
		}
	}
	return req
}

type startAsyncSearchResponse struct {
	SearchID string `json:"search_id" format:"uuid"`
} //	@name	seqapi.v1.StartAsyncSearchResponse

func startAsyncSearchResponseFromProto(resp *seqapi.StartAsyncSearchResponse) startAsyncSearchResponse {
	return startAsyncSearchResponse{
		SearchID: resp.SearchId,
	}
}
