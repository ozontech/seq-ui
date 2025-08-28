package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveSearch go doc.
//
//	@Router		/dashboards/v1/search [post]
//	@ID			dashboards_v1_search
//	@Tags		dashboards_v1
//	@Param		body	body		searchRequest	true	"Request body"
//	@Success	200		{object}	searchResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveSearch(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_search")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq searchRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(httpReq.Limit),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(httpReq.Offset),
		},
	}
	if httpReq.Filter != nil {
		filter, _ := json.Marshal(httpReq.Filter)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "filter",
			Value: attribute.StringValue(string(filter)),
		})
	}

	span.SetAttributes(spanAttributes...)

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.SearchDashboardsRequest{
		Query:  httpReq.Query,
		Limit:  httpReq.Limit,
		Offset: httpReq.Offset,
	}
	if httpReq.Filter != nil {
		req.Filter = &types.SearchDashboardsFilter{
			OwnerName: httpReq.Filter.OwnerName,
		}
	}

	dis, err := a.service.SearchDashboards(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(searchResponse{
		Dashboards: newInfosWithOwner(dis),
	})
}

type searchFilter struct {
	OwnerName *string `json:"owner_name,omitempty"`
} // @name dashboards.v1.SearchFilter

type searchRequest struct {
	Query  string        `json:"query"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Filter *searchFilter `json:"filter,omitempty"`
} // @name dashboards.v1.SearchRequest

type searchResponse struct {
	Dashboards infosWithOwner `json:"dashboards"`
} // @name dashboards.v1.SearchResponse
