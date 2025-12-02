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

// serveFetchAsyncSearchResult go doc.
//
//	@Router		/seqapi/v1/async_search/fetch [post]
//	@ID			seqapi_v1_fetch_async_search_result
//	@Tags		seqapi_v1
//	@Param		body	body		fetchAsyncSearchResultRequest	true	"Request body"
//	@Success	200		{object}	fetchAsyncSearchResultResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error					"An unexpected error response"
//	@Security	bearer
func (a *API) serveFetchAsyncSearchResult(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	if a.asyncSearches == nil {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_fetch_async_search_result")
	defer span.End()

	var httpReq fetchAsyncSearchResultRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse search request: %w", err), http.StatusBadRequest)
		return
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "search_id",
			Value: attribute.StringValue(httpReq.SearchID),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(httpReq.Offset)),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(httpReq.Limit)),
		},
		{
			Key:   "order",
			Value: attribute.StringValue(string(httpReq.Order)),
		},
	}
	span.SetAttributes(spanAttributes...)

	if err := checkUUID(httpReq.SearchID); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if err := checkLimitOffset(httpReq.Limit, httpReq.Offset); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	resp, err := a.asyncSearches.FetchAsyncSearchResult(ctx, httpReq.toProto())
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}
	wr.WriteJson(fetchAsyncSearchResultResponseFromProto(resp))
}

type fetchAsyncSearchResultRequest struct {
	SearchID string `json:"search_id" format:"uuid"`
	Limit    int32  `json:"limit" format:"int32"`
	Offset   int32  `json:"offset" format:"int32"`
	Order    order  `json:"order" default:"desc"`
} //	@name	seqapi.v1.FetchAsyncSearchResultRequest

func (r fetchAsyncSearchResultRequest) toProto() *seqapi.FetchAsyncSearchResultRequest {
	return &seqapi.FetchAsyncSearchResultRequest{
		SearchId: r.SearchID,
		Limit:    r.Limit,
		Offset:   r.Offset,
		Order:    r.Order.toProto(),
	}
}

type fetchAsyncSearchResultResponse struct {
	Status     asyncSearchStatus       `json:"status"`
	Request    startAsyncSearchRequest `json:"request"`
	Response   searchResponse          `json:"response"`
	StartedAt  time.Time               `json:"started_at" format:"date-time"`
	ExpiresAt  time.Time               `json:"expires_at" format:"date-time"`
	CanceledAt *time.Time              `json:"canceled_at,omitempty" format:"date-time"`
	Progress   float64                 `json:"progress"`
	DiskUsage  string                  `json:"disk_usage" format:"int64"`
	Meta       string                  `json:"meta"`
} //	@name	seqapi.v1.FetchAsyncSearchResultResponse

func fetchAsyncSearchResultResponseFromProto(resp *seqapi.FetchAsyncSearchResultResponse) fetchAsyncSearchResultResponse {
	var canceledAt *time.Time
	if resp.CanceledAt != nil {
		t := resp.CanceledAt.AsTime()
		canceledAt = &t
	}

	return fetchAsyncSearchResultResponse{
		Status:     asyncSearchStatusFromProto(resp.Status),
		Request:    startAsyncSearchRequestFromProto(resp.Request),
		Response:   searchResponseFromProto(resp.Response, true),
		StartedAt:  resp.StartedAt.AsTime(),
		ExpiresAt:  resp.ExpiresAt.AsTime(),
		CanceledAt: canceledAt,
		Progress:   resp.Progress,
		DiskUsage:  strconv.FormatUint(resp.DiskUsage, 10),
		Meta:       resp.Meta,
	}
}
