package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveRestore go doc.
//
//	@Router		/massexport/v1/restore [post]
//	@ID			massexport_v1_restore
//	@Tags		massexport_v1
//	@Param		body	body		restoreRequest	true	"Request body"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveRestore(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "massexport_v1_restore")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq restoreRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("parse coninue export request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(httpReq.SessionID),
	})

	err := a.exporter.RestoreExport(ctx, httpReq.SessionID)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type restoreRequest struct {
	SessionID string `json:"session_id"`
} // @name massexport.v1.RestoreRequest
