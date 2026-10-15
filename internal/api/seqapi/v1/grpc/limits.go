package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetLimits(ctx context.Context, _ *seqapi.GetLimitsRequest) (*seqapi.GetLimitsResponse, error) {
	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &seqapi.GetLimitsResponse{
		MaxSearchLimit:            params.options.Limits.MaxSearchLimit,
		MaxExportLimit:            params.options.Limits.MaxExportLimit,
		MaxParallelExportRequests: int32(params.options.Limits.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(params.options.Limits.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(params.options.Limits.SeqCLIMaxSearchLimit),
	}, nil
}
