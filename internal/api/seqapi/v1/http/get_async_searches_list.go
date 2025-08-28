package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetAsyncSearchesList go doc.
//
//	@Router		/seqapi/v1/async_search/list [post]
//	@ID			seqapi_v1_get_async_searches_list
//	@Tags		seqapi_v1
//	@Param		body	body		getAsyncSearchesListRequest		true	"Request body"
//	@Success	200		{object}	getAsyncSearchesListResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error					"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetAsyncSearchesList(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	if a.asyncSearches == nil {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_async_searches_list")
	defer span.End()

	var httpReq getAsyncSearchesListRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse search request: %w", err), http.StatusBadRequest)
		return
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "offset",
			Value: attribute.IntValue(int(httpReq.Offset)),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(httpReq.Limit)),
		},
	}
	if httpReq.Status != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "status",
			Value: attribute.StringValue(string(*httpReq.Status)),
		})
	}
	if httpReq.Owner != nil {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "owner",
			Value: attribute.StringValue(*httpReq.Owner),
		})
	}

	span.SetAttributes(spanAttributes...)

	if err := checkLimitOffset(httpReq.Limit, httpReq.Offset); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	parsedReq, err := httpReq.toProto()
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	resp, err := a.asyncSearches.GetAsyncSearchesList(ctx, parsedReq)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(getAsyncSearchesListResponseFromProto(resp))
}

type getAsyncSearchesListRequest struct {
	Status *asyncSearchStatus `json:"status"`
	Limit  int32              `json:"limit" format:"int32"`
	Offset int32              `json:"offset" format:"int32"`
	Owner  *string            `json:"owner_name"`
} //	@name	seqapi.v1.GetAsyncSearchesListRequest

func (r getAsyncSearchesListRequest) toProto() (*seqapi.GetAsyncSearchesListRequest, error) {
	var status *seqapi.AsyncSearchStatus
	if r.Status != nil {
		protoStatus, err := asyncSearchStatusToProto(*r.Status)
		if err != nil {
			return nil, err
		}
		status = &protoStatus
	}

	return &seqapi.GetAsyncSearchesListRequest{
		Status:    status,
		Limit:     r.Limit,
		Offset:    r.Offset,
		OwnerName: r.Owner,
	}, nil
}

type getAsyncSearchesListResponse struct {
	Searches []asyncSearchesListItem `json:"searches"`
} //	@name	seqapi.v1.GetAsyncSearchesListResponse

type asyncSearchesListItem struct {
	SearchID   string                  `json:"search_id" format:"uuid"`
	Status     asyncSearchStatus       `json:"status"`
	Request    startAsyncSearchRequest `json:"request"`
	StartedAt  time.Time               `json:"started_at" format:"date-time"`
	ExpiresAt  time.Time               `json:"expires_at" format:"date-time"`
	CanceledAt *time.Time              `json:"canceled_at,omitempty" format:"date-time"`
	Progress   float64                 `json:"progress"`
	DiskUsage  string                  `json:"disk_usage" format:"int64"`
	OwnerName  string                  `json:"owner_name"`
} //	@name	seqapi.v1.AsyncSearchesListItem

func getAsyncSearchesListResponseFromProto(resp *seqapi.GetAsyncSearchesListResponse) getAsyncSearchesListResponse {
	searches := make([]asyncSearchesListItem, 0, len(resp.Searches))

	for _, s := range resp.Searches {
		var canceledAt *time.Time
		if s.CanceledAt != nil {
			t := s.CanceledAt.AsTime()
			canceledAt = &t
		}

		searches = append(searches, asyncSearchesListItem{
			SearchID:   s.SearchId,
			Status:     asyncSearchStatusFromProto(s.Status),
			Request:    startAsyncSearchRequestFromProto(s.Request),
			StartedAt:  s.StartedAt.AsTime(),
			ExpiresAt:  s.ExpiresAt.AsTime(),
			CanceledAt: canceledAt,
			Progress:   s.Progress,
			DiskUsage:  strconv.FormatUint(s.DiskUsage, 10),
			OwnerName:  s.OwnerName,
		})
	}

	return getAsyncSearchesListResponse{
		Searches: searches,
	}
}

func startAsyncSearchRequestFromProto(r *seqapi.StartAsyncSearchRequest) startAsyncSearchRequest {
	var hist *AsyncSearchRequestHistogram
	if r.Hist != nil {
		hist = &AsyncSearchRequestHistogram{
			Interval: r.Hist.Interval,
		}
	}

	return startAsyncSearchRequest{
		Retention:    r.Retention.String(),
		Query:        r.Query,
		From:         r.From.AsTime(),
		To:           r.To.AsTime(),
		Aggregations: aggregationQueriesFromProto(r.Aggs),
		Histogram:    hist,
		WithDocs:     r.WithDocs,
	}
}

func aggregationQueriesFromProto(aggs []*seqapi.AggregationQuery) aggregationQueries {
	result := make(aggregationQueries, 0, len(aggs))

	for _, agg := range aggs {
		result = append(result, aggregationQuery{
			Field:     agg.Field,
			GroupBy:   agg.GroupBy,
			Func:      aggregationFuncFromProto(agg.Func),
			Quantiles: agg.Quantiles,
		})
	}

	return result
}
