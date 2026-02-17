package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveCheck go doc.
//
//	@Router		/massexport/v1/check [post]
//	@ID			massexport_v1_check
//	@Tags		massexport_v1
//	@Param		body	body		checkRequest	true	"Request body"
//	@Success	200		{object}	checkResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "massexport_v1_check")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq checkRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("parse check export request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "session_id",
		Value: attribute.StringValue(httpReq.SessionID),
	})

	info, err := a.exporter.CheckExport(ctx, httpReq.SessionID)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(convertExportInfo(info))
}

type checkRequest struct {
	SessionID string `json:"session_id"`
} //	@name	massexport.v1.CheckRequest

type checkResponse struct {
	ID           string       `json:"id"`
	Status       exportStatus `json:"status"`
	Progress     float64      `json:"progress"`
	UserID       string       `json:"user_id"`
	Links        string       `json:"links"`
	StartedAt    time.Time    `json:"started_at"`
	FinishedAt   time.Time    `json:"finished_at"`
	Duration     string       `json:"duration"`
	Error        string       `json:"error"`
	UnpackedSize int          `json:"unpacked_size"`
	PackedSize   int          `json:"packed_size"`
} //	@name	massexport.v1.CheckResponse

type exportStatus string //	@name	massexport.v1.ExportStatus

const (
	exportStatusUnspecified exportStatus = "UNSPECIFIED"
	exportStatusStart       exportStatus = "START"
	exportStatusCancel      exportStatus = "CANCEL"
	exportStatusFail        exportStatus = "FAIL"
	exportStatusFinish      exportStatus = "FINISH"
)

func convertExportStatus(s types.ExportStatus) exportStatus {
	switch s {
	case types.ExportStatusUnspecified:
		return exportStatusUnspecified
	case types.ExportStatusStart:
		return exportStatusStart
	case types.ExportStatusCancel:
		return exportStatusCancel
	case types.ExportStatusFail:
		return exportStatusFail
	case types.ExportStatusFinish:
		return exportStatusFinish
	default:
		panic(fmt.Sprintf("unknown export status: %s", s.String()))
	}
}

func convertExportInfo(info types.ExportInfo) checkResponse {
	var duration time.Duration
	if !info.StartedAt.IsZero() && !info.FinishedAt.IsZero() {
		duration = info.FinishedAt.Sub(info.StartedAt)
	}

	return checkResponse{
		ID:           info.ID,
		Status:       convertExportStatus(info.Status),
		Progress:     info.Progress,
		Links:        info.Links,
		UserID:       info.UserID,
		StartedAt:    info.StartedAt,
		FinishedAt:   info.FinishedAt,
		Duration:     duration.String(),
		Error:        info.Error,
		UnpackedSize: info.TotalSize.Unpacked,
		PackedSize:   info.TotalSize.Packed,
	}
}
