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

	_, options, err := a.GetClientFromEnv(env)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}
	// Тут делема, как нам связать grpc get_limits и http get_limits, бизнес-логику продумывать?
	// md := metadata.New(map[string]string{
	// 	"env": env,
	// })
	// grpcCtx := metadata.NewOutgoingContext(ctx, md)

	httputil.NewWriter(w).WriteJson(getLimitsResponse{
		MaxSearchLimit:            options.MaxSearchLimit,
		MaxExportLimit:            options.MaxExportLimit,
		MaxParallelExportRequests: int32(options.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(options.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(options.SeqCLIMaxSearchLimit),
	})
}

type getLimitsResponse struct {
	MaxSearchLimit            int32 `json:"maxSearchLimit" format:"int32"`
	MaxExportLimit            int32 `json:"maxExportLimit" format:"int32"`
	MaxParallelExportRequests int32 `json:"maxParallelExportRequests" format:"int32"`
	MaxAggregationsPerRequest int32 `json:"maxAggregationsPerRequest" format:"int32"`
	SeqCliMaxSearchLimit      int32 `json:"seqCliMaxSearchLimit" format:"int32"`
} //	@name	seqapi.v1.GetLimitsResponse
