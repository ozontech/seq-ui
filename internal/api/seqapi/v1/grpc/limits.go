package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetLimits(ctx context.Context, _ *seqapi.GetLimitsRequest) (*seqapi.GetLimitsResponse, error) {
	params := a.paramsByEnv[types.GetEnvKey(ctx)]

	return &seqapi.GetLimitsResponse{
		MaxSearchLimit:            params.options.Limits.MaxSearchLimit,
		MaxExportLimit:            params.options.Limits.MaxExportLimit,
		MaxParallelExportRequests: int32(params.options.Limits.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(params.options.Limits.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(params.options.Limits.SeqCLIMaxSearchLimit),
	}, nil
}
