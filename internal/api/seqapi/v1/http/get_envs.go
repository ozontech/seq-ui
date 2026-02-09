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
//	@Security	bearer
func (a *API) serveGetEnvs(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	// Заглушка, чтобы линтер не ругался (type getEnvsResponse is unused)
	envs := []EnvInfo{
		{
			Env:            "Окружение балдежа",
			MaxSearchLimit: 1,
			MaxExportLimit: 1,
		},
	}

	response := getEnvsResponse{Envs: envs}
	wr.WriteJson(response)
}

type getEnvsResponse struct {
	Envs []EnvInfo `json:"envs"`
} //	@name	getEnvsResponse

type EnvInfo struct {
	Env            string `json:"env"`
	MaxSearchLimit uint32 `json:"max_search_limit"`
	MaxExportLimit uint32 `json:"max_export_limit"`
}
