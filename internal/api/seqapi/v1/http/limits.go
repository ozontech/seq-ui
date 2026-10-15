package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetLimits go doc.
//
//	@Router		/seqapi/v1/limits [get]
//	@ID			seqapi_v1_get_limits
//	@Tags		seqapi_v1
//	@Param		env		query		string				false	"Environment"
//	@Success	200		{object}	getLimitsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetLimits(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_limits")
	defer span.End()

	wr := httputil.NewWriter(w)
	env := types.GetEnvKey(ctx)

	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	httputil.NewWriter(w).WriteJson(getLimitsResponse{
		MaxSearchLimit:            int(params.options.Limits.MaxSearchLimit),
		MaxExportLimit:            int(params.options.Limits.MaxExportLimit),
		MaxParallelExportRequests: params.options.Limits.MaxParallelExportRequests,
		MaxAggregationsPerRequest: params.options.Limits.MaxAggregationsPerRequest,
		SeqCliMaxSearchLimit:      params.options.Limits.SeqCLIMaxSearchLimit,
	})
}

type getLimitsResponse struct {
	MaxSearchLimit            int `json:"maxSearchLimit"`
	MaxExportLimit            int `json:"maxExportLimit"`
	MaxParallelExportRequests int `json:"maxParallelExportRequests"`
	MaxAggregationsPerRequest int `json:"maxAggregationsPerRequest"`
	SeqCliMaxSearchLimit      int `json:"seqCliMaxSearchLimit"`
} //	@name	seqapi.v1.GetLimitsResponse
