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

// serveCreate go doc.
//
//	@Router		/dashboards/v1/ [post]
//	@ID			dashboards_v1_create
//	@Tags		dashboards_v1
//	@Param		body	body		createRequest	true	"Request body"
//	@Success	200		{object}	createResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveCreate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_create")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq createRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(httpReq.Name),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.CreateDashboardRequest{
		ProfileID: profileID,
		Name:      httpReq.Name,
		Meta:      httpReq.Meta,
	}
	uuid, err := a.service.CreateDashboard(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(createResponse{UUID: uuid})
}

type createRequest struct {
	Name string `json:"name"`
	Meta string `json:"meta"`
} // @name dashboards.v1.CreateRequest

type createResponse struct {
	UUID string `json:"uuid"`
} // @name dashboards.v1.CreateResponse
