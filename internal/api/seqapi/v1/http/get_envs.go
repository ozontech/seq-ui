package http

import (
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
)

// serveGetEnvs go doc.
//
//	@Router		/seqapi/v1/envs [get]
//	@ID			seqapi_v1_get_envs
//	@Tags		seqapi_v1
//	@Produce	json
//	@Success	200		{object}	getEnvsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
func (a *API) serveGetEnvs(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)
	wr.WriteJson(getEnvsResponse{})
}

//nolint:unused
type getEnvsResponse struct {
	Envs []envInfo `json:"envs"`
} // @name seqapi.v1.GetEnvsResponse

//nolint:unused
type envInfo struct {
	Env                       string `json:"env"`
	MaxSearchLimit            uint32 `json:"max_search_limit"`
	MaxExportLimit            uint32 `json:"max_export_limit"`
	MaxParallelExportRequests uint32 `json:"max_parallel_export_requests"`
	MaxAggregationsPerRequest uint32 `json:"max_aggregations_per_request"`
	SeqCliMaxSearchLimit      uint32 `json:"seq_cli_max_search_limit"`
} // @name seqapi.v1.EnvInfo
