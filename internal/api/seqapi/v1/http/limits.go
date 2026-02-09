package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
)

// serveGetLimits go doc.
//
//	@Router		/seqapi/v1/limits [get]
//	@ID			seqapi_v1_getLimits
//	@Tags		seqapi_v1
//	@Param		env		query		string				true	"Environment"
//	@Success	200		{object}	getLimitsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetLimits(w http.ResponseWriter, _ *http.Request) {
	httputil.NewWriter(w).WriteJson(getLimitsResponse{
		MaxSearchLimit:            a.config.MaxSearchLimit,
		MaxExportLimit:            a.config.MaxExportLimit,
		MaxParallelExportRequests: int32(a.config.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(a.config.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(a.config.SeqCLIMaxSearchLimit),
	})
}

type getLimitsResponse struct {
	MaxSearchLimit            int32 `json:"maxSearchLimit" format:"int32"`
	MaxExportLimit            int32 `json:"maxExportLimit" format:"int32"`
	MaxParallelExportRequests int32 `json:"maxParallelExportRequests" format:"int32"`
	MaxAggregationsPerRequest int32 `json:"maxAggregationsPerRequest" format:"int32"`
	SeqCliMaxSearchLimit      int32 `json:"seqCliMaxSearchLimit" format:"int32"`
} //	@name	seqapi.v1.GetLimitsResponse
