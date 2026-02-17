package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func (a *API) GetEnvs(_ context.Context, _ *seqapi.GetEnvsRequest) (*seqapi.GetEnvsResponse, error) {
	protoEnvs := make([]*seqapi.GetEnvsResponse_Env, 0, len(a.config.Envs))
	for envName, envConfig := range a.config.Envs {
		protoEnv := &seqapi.GetEnvsResponse_Env{
			Env:                       envName,
			MaxSearchLimit:            uint32(envConfig.Options.MaxSearchLimit),
			MaxExportLimit:            uint32(envConfig.Options.MaxExportLimit),
			MaxParallelExportRequests: uint32(envConfig.Options.MaxParallelExportRequests),
			MaxAggregationsPerRequest: uint32(envConfig.Options.MaxAggregationsPerRequest),
			SeqCliMaxSearchLimit:      uint32(envConfig.Options.SeqCLIMaxSearchLimit),
		}
		protoEnvs = append(protoEnvs, protoEnv)
	}
	return &seqapi.GetEnvsResponse{
		Envs: protoEnvs,
	}, nil
}
