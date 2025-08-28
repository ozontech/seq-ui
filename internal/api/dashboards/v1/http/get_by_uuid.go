package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveGetByUUID go doc.
//
//	@Router		/dashboards/v1/{uuid} [get]
//	@ID			dashboards_v1_getByUUID
//	@Tags		dashboards_v1
//	@Param		uuid	path		string			true	"Dashboard UUID"
//	@Success	200		{object}	dashboard		"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetByUUID(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_get_by_uuid")
	defer span.End()

	wr := httputil.NewWriter(w)

	uuid := chi.URLParam(r, "uuid")

	span.SetAttributes(attribute.KeyValue{
		Key:   "uuid",
		Value: attribute.StringValue(uuid),
	})

	// check auth and create profile if its doesn't exist
	if _, err := a.profiles.GeIDFromContext(ctx); err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	d, err := a.service.GetDashboardByUUID(ctx, uuid)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(newDashboard(d))
}
