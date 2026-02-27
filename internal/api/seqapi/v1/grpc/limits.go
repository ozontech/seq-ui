package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *API) GetLimits(ctx context.Context, _ *seqapi.GetLimitsRequest) (*seqapi.GetLimitsResponse, error) {
	env := a.GetEnvFromContext(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &seqapi.GetLimitsResponse{
		MaxSearchLimit:            params.options.MaxSearchLimit,
		MaxExportLimit:            params.options.MaxExportLimit,
		MaxParallelExportRequests: int32(params.options.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(params.options.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(params.options.SeqCLIMaxSearchLimit),
	}, nil
}
