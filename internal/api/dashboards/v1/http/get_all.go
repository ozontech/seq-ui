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

// serveGetAll go doc.
//
//	@Router		/dashboards/v1/all [post]
//	@ID			dashboards_v1_getAll
//	@Tags		dashboards_v1
//	@Param		body	body		getAllRequest	true	"Request body"
//	@Success	200		{object}	getAllResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetAll(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_get_all")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getAllRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(httpReq.Limit),
		},
		attribute.KeyValue{
			Key:   "offset",
			Value: attribute.IntValue(httpReq.Offset),
		},
	)

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.GetAllDashboardsRequest{
		Limit:  httpReq.Limit,
		Offset: httpReq.Offset,
	}
	dis, err := a.service.GetAllDashboards(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getAllResponse{
		Dashboards: newInfosWithOwner(dis),
	})
}

type getAllRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
} // @name dashboards.v1.GetAllRequest

type getAllResponse struct {
	Dashboards infosWithOwner `json:"dashboards"`
} // @name dashboards.v1.GetAllResponse
