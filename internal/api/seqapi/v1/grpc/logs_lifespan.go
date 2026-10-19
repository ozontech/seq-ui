package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/lifespan"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetLogsLifespan(ctx context.Context, _ *seqapi.GetLogsLifespanRequest) (*seqapi.GetLogsLifespanResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_logs_lifespan")
	defer span.End()

	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if countStr, ok := params.caches.LogsLifespan.Get(ctx, ""); ok {
		count, err := strconv.Atoi(countStr)
		if err == nil {
			res := time.Duration(count) * lifespan.MeasureUnit
			logger.Debug("got logs lifespan from cache")
			return &seqapi.GetLogsLifespanResponse{Lifespan: durationpb.New(res)}, nil
		}
		logger.Error("failed to parse logs lifespan", zap.Error(err))
	}

	clientStatus, err := params.client.Status(ctx, &seqapi.StatusRequest{})
	if err != nil {
		return nil, grpcutil.ProcessError(fmt.Errorf("get status: %w", err))
	}

	if clientStatus.OldestStorageTime == nil {
		return nil, grpcutil.ProcessError(errors.New("oldest timestamp is nil"))
	}

	count := int(a.nowFn().Sub(clientStatus.OldestStorageTime.AsTime()) / lifespan.MeasureUnit)
	res := time.Duration(count) * lifespan.MeasureUnit

	params.caches.LogsLifespan.Set(ctx, "", strconv.Itoa(count))

	logger.Debug("got logs lifespan from seq-proxy")
	return &seqapi.GetLogsLifespanResponse{Lifespan: durationpb.New(res)}, nil
}
