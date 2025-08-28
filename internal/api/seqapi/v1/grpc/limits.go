package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetLimits(_ context.Context, _ *seqapi.GetLimitsRequest) (*seqapi.GetLimitsResponse, error) {
	return &seqapi.GetLimitsResponse{
		MaxSearchLimit:            a.config.MaxSearchLimit,
		MaxExportLimit:            a.config.MaxExportLimit,
		MaxParallelExportRequests: int32(a.config.MaxParallelExportRequests),
		MaxAggregationsPerRequest: int32(a.config.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      int32(a.config.SeqCLIMaxSearchLimit),
	}, nil
}
