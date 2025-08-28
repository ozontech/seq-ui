package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveDelete go doc.
//
//	@Router		/dashboards/v1/{uuid} [delete]
//	@ID			dashboards_v1_delete
//	@Tags		dashboards_v1
//	@Param		uuid	path		string			true	"Dashboard UUID"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveDelete(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_delete")
	defer span.End()

	wr := httputil.NewWriter(w)

	uuid := chi.URLParam(r, "uuid")

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "uuid",
			Value: attribute.StringValue(uuid),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.DeleteDashboardRequest{
		UUID:      uuid,
		ProfileID: profileID,
	}
	err = a.service.DeleteDashboard(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
