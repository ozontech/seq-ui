package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveCancel go doc.
//
//	@Router		/massexport/v1/cancel [post]
//	@ID			massexport_v1_cancel
//	@Tags		massexport_v1
//	@Param		body	body		cancelRequest	true	"Request body"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveCancel(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "massexport_v1_cancel")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq cancelRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("parse cancel export request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(httpReq.SessionID),
	})

	err := a.exporter.CancelExport(ctx, httpReq.SessionID)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type cancelRequest struct {
	SessionID string `json:"session_id"`
} // @name massexport.v1.CancelRequest
