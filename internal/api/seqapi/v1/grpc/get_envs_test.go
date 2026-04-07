package grpc

import (
	"context"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGetEnvs(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Handlers
		want *seqapi.GetEnvsResponse
	}{
		{
			name: "single_env",
			cfg: config.Handlers{
				SeqAPI: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCLIMaxSearchLimit:      10000,
					},
				},
			},
			want: &seqapi.GetEnvsResponse{
				Envs: []*seqapi.GetEnvsResponse_Env{
					{
						Env:                       "",
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      10000,
					},
				},
			},
		},
		{
			name: "multiple_envs",
			cfg: config.Handlers{
				SeqAPI: config.SeqAPI{
					Envs: map[string]config.SeqAPIEnv{
						"test": {
							Options: &config.SeqAPIOptions{
								MaxSearchLimit:            100,
								MaxExportLimit:            200,
								MaxParallelExportRequests: 2,
								MaxAggregationsPerRequest: 5,
								SeqCLIMaxSearchLimit:      10000,
							},
						},
						"prod": {
							Options: &config.SeqAPIOptions{
								MaxSearchLimit:            150,
								MaxExportLimit:            250,
								MaxParallelExportRequests: 3,
								MaxAggregationsPerRequest: 6,
								SeqCLIMaxSearchLimit:      15000,
							},
						},
						"cluster-10": {
							Options: &config.SeqAPIOptions{
								MaxSearchLimit:            150,
								MaxExportLimit:            250,
								MaxParallelExportRequests: 3,
								MaxAggregationsPerRequest: 6,
								SeqCLIMaxSearchLimit:      15000,
							},
						},
						"cluster-102": {
							Options: &config.SeqAPIOptions{
								MaxSearchLimit:            100,
								MaxExportLimit:            200,
								MaxParallelExportRequests: 2,
								MaxAggregationsPerRequest: 5,
								SeqCLIMaxSearchLimit:      10000,
							},
						},
						"cluster-220": {
							Options: &config.SeqAPIOptions{
								MaxSearchLimit:            100,
								MaxExportLimit:            200,
								MaxParallelExportRequests: 2,
								MaxAggregationsPerRequest: 5,
								SeqCLIMaxSearchLimit:      10000,
							},
						},
					},
					DefaultEnv: "test",
				},
			},
			want: &seqapi.GetEnvsResponse{
				Envs: []*seqapi.GetEnvsResponse_Env{
					{
						Env:                       "cluster-10",
						MaxSearchLimit:            150,
						MaxExportLimit:            250,
						MaxParallelExportRequests: 3,
						MaxAggregationsPerRequest: 6,
						SeqCliMaxSearchLimit:      15000,
					},
					{
						Env:                       "cluster-102",
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      10000,
					},
					{
						Env:                       "cluster-220",
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      10000,
					},
					{
						Env:                       "prod",
						MaxSearchLimit:            150,
						MaxExportLimit:            250,
						MaxParallelExportRequests: 3,
						MaxAggregationsPerRequest: 6,
						SeqCliMaxSearchLimit:      15000,
					},
					{
						Env:                       "test",
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      10000,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := API{
				config:       tt.cfg,
				envsResponse: parseEnvs(tt.cfg.SeqAPI),
			}
			resp, err := api.GetEnvs(context.TODO(), &seqapi.GetEnvsRequest{})
			require.NoError(t, err)
			require.True(t, proto.Equal(tt.want, resp))
		})
	}
}
