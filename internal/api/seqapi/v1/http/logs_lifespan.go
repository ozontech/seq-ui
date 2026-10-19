package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/lifespan"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetLogsLifespan go doc.
//
//	@Router		/seqapi/v1/logs_lifespan [get]
//	@ID			seqapi_v1_get_logs_lifespan
//	@Tags		seqapi_v1
//	@Param		env		query		string					false	"Environment"
//	@Success	200		{object}	getLogsLifespanResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
func (a *API) serveGetLogsLifespan(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_logs_lifespan")
	defer span.End()

	wr := httputil.NewWriter(w)

	env := types.GetEnvKey(ctx)
	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if resStr, ok := params.caches.LogsLifespan.Get(ctx, ""); ok {
		res, err := strconv.Atoi(resStr)
		if err == nil {
			logger.Debug("got logs lifespan from cache")
			wr.WriteJson(getLogsLifespanResponse{Lifespan: res})
			return
		}
		logger.Error("failed to parse logs lifespan", zap.Error(err))
	}

	clientStatus, err := params.client.Status(ctx, &seqapi.StatusRequest{})
	if err != nil {
		wr.Error(fmt.Errorf("get status: %w", err), http.StatusInternalServerError)
		return
	}

	if clientStatus.OldestStorageTime == nil {
		wr.Error(errors.New("oldest timestamp is nil"), http.StatusInternalServerError)
		return
	}

	res := int(a.nowFn().Sub(clientStatus.OldestStorageTime.AsTime()) / lifespan.MeasureUnit)

	params.caches.LogsLifespan.Set(ctx, "", strconv.Itoa(res))

	logger.Debug("got logs lifespan from seq-proxy")
	wr.WriteJson(getLogsLifespanResponse{Lifespan: res})
}

type getLogsLifespanResponse struct {
	Lifespan int `json:"lifespan"`
} //	@name	seqapi.v1.GetLogsLifespanResponse
