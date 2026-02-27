package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetLimits go doc.
//
//	@Router		/seqapi/v1/limits [get]
//	@ID			seqapi_v1_getLimits
//	@Tags		seqapi_v1
//	@Param		env		query		string				false	"Environment"
//	@Success	200		{object}	getLimitsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetLimits(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_limits")
	defer span.End()

	wr := httputil.NewWriter(w)
	env := getEnvFromContext(ctx)

	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	httputil.NewWriter(w).WriteJson(getLimitsResponse{
		MaxSearchLimit:            params.options.MaxSearchLimit,
		MaxExportLimit:            params.options.MaxExportLimit,
		MaxParallelExportRequests: int32(params.options.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(params.options.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(params.options.SeqCLIMaxSearchLimit),
	})
}

type getLimitsResponse struct {
	MaxSearchLimit            int32 `json:"maxSearchLimit" format:"int32"`
	MaxExportLimit            int32 `json:"maxExportLimit" format:"int32"`
	MaxParallelExportRequests int32 `json:"maxParallelExportRequests" format:"int32"`
	MaxAggregationsPerRequest int32 `json:"maxAggregationsPerRequest" format:"int32"`
	SeqCliMaxSearchLimit      int32 `json:"seqCliMaxSearchLimit" format:"int32"`
} //	@name	seqapi.v1.GetLimitsResponse
