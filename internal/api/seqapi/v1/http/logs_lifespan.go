package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/lifespan"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.uber.org/zap"
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

	env := getEnvFromContext(ctx)
	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	cacheKey := params.options.LogsLifespanCacheKey

	if resStr, err := a.redisCache.Get(ctx, cacheKey); err == nil {
		res := 0
		res, err = strconv.Atoi(resStr)
		if err == nil {
			logger.Debug("got logs lifespan from cache")
			wr.WriteJson(getLogsLifespanResponse{Lifespan: int64(res)})
			return
		}

		logger.Error("can't parse logs lifespan", zap.Error(err))
	} else if !errors.Is(err, cache.ErrNotFound) {
		logger.Error("can't get logs lifespan from cache", zap.Error(err))
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

	err = a.redisCache.SetWithTTL(ctx, cacheKey, strconv.Itoa(res), params.options.LogsLifespanCacheTTL)
	if err != nil {
		logger.Error("can't set logs lifespan to cache", zap.Error(err))
	}

	logger.Debug("got logs lifespan from seq-proxy")
	wr.WriteJson(getLogsLifespanResponse{Lifespan: int64(res)})
}

type getLogsLifespanResponse struct {
	Lifespan int64 `json:"lifespan"`
} //	@name	seqapi.v1.GetLogsLifespanResponse
