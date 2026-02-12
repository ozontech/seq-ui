package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/massexport/v1/util"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveStart go doc.
//
//	@Router		/massexport/v1/start [post]
//	@ID			massexport_v1_start
//	@Tags		massexport_v1
//	@Param		body	body		startRequest	true	"Request body"
//	@Success	200		{object}	startResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveStart(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "massexport_v1_start")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq startRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("parse start export request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes([]attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		{
			Key:   "from",
			Value: attribute.StringValue(httpReq.From.Format(time.RFC3339)),
		},
		{
			Key:   "to",
			Value: attribute.StringValue(httpReq.To.Format(time.RFC3339)),
		},
		{
			Key:   "window",
			Value: attribute.StringValue(httpReq.Window),
		},
		{
			Key:   "name",
			Value: attribute.StringValue(httpReq.Name),
		},
	}...)

	if httpReq.Name == "" {
		wr.Error(errors.New("empty export name"), http.StatusBadRequest)
		return
	}

	window, err := util.ParseWindow(httpReq.Window)
	if err != nil {
		wr.Error(fmt.Errorf("parse 'window': %w", err), http.StatusBadRequest)
		return
	}

	if !(httpReq.From.Before(httpReq.To)) {
		wr.Error(errors.New("'from' is not before 'to'"), http.StatusBadRequest)
		return
	}

	if window > httpReq.To.Sub(httpReq.From) {
		wr.Error(errors.New("'window' is larger than whole interval"), http.StatusBadRequest)
		return
	}

	result, err := a.exporter.StartExport(ctx, types.StartExportRequest{
		Query:  httpReq.Query,
		From:   httpReq.From,
		To:     httpReq.To,
		Window: window,
		Name:   httpReq.Name,
	})

	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(startResponse{
		SessionID: result.SessionID,
	})
}

type startRequest struct {
	Query  string    `json:"query"`
	From   time.Time `json:"from" format:"date-time"`
	To     time.Time `json:"to" format:"date-time"`
	Window string    `json:"window"`
	Name   string    `json:"name"`
} //	@name	massexport.v1.StartRequest

type startResponse struct {
	SessionID string `json:"session_id"`
} //	@name	massexport.v1.StartResponse
