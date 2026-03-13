package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/tracing"
)

// serveJobs go doc.
//
//	@Router		/massexport/v1/jobs [get]
//	@ID			massexport_v1_jobs
//	@Tags		massexport_v1
//	@Success	200		{object}	getAllResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveJobs(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "massexport_v1_jobs")
	defer span.End()

	wr := httputil.NewWriter(w)

	exports, err := a.exporter.GetAll(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	result := make([]checkResponse, 0, len(exports))
	for i := range exports {
		result = append(result, convertExportInfo(exports[i]))
	}

	wr.WriteJson(getAllResponse{Exports: result})
}

type getAllResponse struct {
	Exports []checkResponse `json:"exports"`
} //	@name	massexport.v1.GetAllResponse
