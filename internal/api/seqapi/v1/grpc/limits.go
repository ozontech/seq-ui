package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetLimits(ctx context.Context, _ *seqapi.GetLimitsRequest) (*seqapi.GetLimitsResponse, error) {
	env, err := a.GetEnvFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, options, err := a.GetClientFromEnv(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &seqapi.GetLimitsResponse{
		MaxSearchLimit:            options.MaxSearchLimit,
		MaxExportLimit:            options.MaxExportLimit,
		MaxParallelExportRequests: int32(options.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(options.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(options.SeqCLIMaxSearchLimit),
	}, nil
}
