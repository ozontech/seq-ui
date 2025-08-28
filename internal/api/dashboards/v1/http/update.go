package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveUpdate go doc.
//
//	@Router		/dashboards/v1/{uuid} [patch]
//	@ID			dashboards_v1_update
//	@Tags		dashboards_v1
//	@Param		uuid	path		string			true	"Dashboard UUID"
//	@Param		body	body		updateRequest	true	"Request body"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_update")
	defer span.End()

	wr := httputil.NewWriter(w)

	uuid := chi.URLParam(r, "uuid")

	var httpReq updateRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "uuid",
			Value: attribute.StringValue(uuid),
		},
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(httpReq.GetName()),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.UpdateDashboardRequest{
		UUID:      uuid,
		ProfileID: profileID,
		Name:      httpReq.Name,
		Meta:      httpReq.Meta,
	}
	err = a.service.UpdateDashboard(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type updateRequest struct {
	Name *string `json:"name"`
	Meta *string `json:"meta"`
} // @name dashboards.v1.UpdateRequest

func (r updateRequest) GetName() string {
	if r.Name != nil {
		return *r.Name
	}
	return ""
}

func (r updateRequest) GetMeta() string {
	if r.Meta != nil {
		return *r.Meta
	}
	return ""
}
