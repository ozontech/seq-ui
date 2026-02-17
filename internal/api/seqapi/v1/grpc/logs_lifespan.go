package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/lifespan"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (a *API) GetLogsLifespan(ctx context.Context, _ *seqapi.GetLogsLifespanRequest) (*seqapi.GetLogsLifespanResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_logs_lifespan")
	defer span.End()

	env, err := a.GetEnvFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	client, options, err := a.GetClientFromEnv(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "env",
		Value: attribute.StringValue(env),
	})

	cacheKey := options.LogsLifespanCacheKey

	if countStr, err := a.redisCache.Get(ctx, cacheKey); err == nil {
		count := 0
		count, err = strconv.Atoi(countStr)
		if err == nil {
			res := time.Duration(count) * lifespan.MeasureUnit
			logger.Debug("got logs lifespan from cache")
			return &seqapi.GetLogsLifespanResponse{Lifespan: durationpb.New(res)}, nil
		}

		logger.Error("can't parse logs lifespan", zap.Error(err))
	} else if !errors.Is(err, cache.ErrNotFound) {
		logger.Error("can't get logs lifespan from cache", zap.Error(err))
	}

	clientStatus, err := client.Status(ctx, &seqapi.StatusRequest{})
	if err != nil {
		return nil, grpcutil.ProcessError(fmt.Errorf("get status: %w", err))
	}

	if clientStatus.OldestStorageTime == nil {
		return nil, grpcutil.ProcessError(errors.New("oldest timestamp is nil"))
	}

	count := int(a.nowFn().Sub(clientStatus.OldestStorageTime.AsTime()) / lifespan.MeasureUnit)
	res := time.Duration(count) * lifespan.MeasureUnit

	err = a.redisCache.SetWithTTL(ctx, cacheKey, strconv.Itoa(count), options.LogsLifespanCacheTTL)
	if err != nil {
		logger.Error("can't set logs lifespan to cache", zap.Error(err))
	}

	logger.Debug("got logs lifespan from seq-proxy")
	return &seqapi.GetLogsLifespanResponse{Lifespan: durationpb.New(res)}, nil
}
