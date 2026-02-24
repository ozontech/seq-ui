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
		cfg  config.SeqAPI
		want *seqapi.GetEnvsResponse
	}{
		{
			name: "single_env",
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:            100,
					MaxExportLimit:            200,
					MaxParallelExportRequests: 2,
					MaxAggregationsPerRequest: 5,
					SeqCLIMaxSearchLimit:      10000,
				},
			},
			want: &seqapi.GetEnvsResponse{
				Envs: []*seqapi.GetEnvsResponse_Env{
					{
						Env:                       "default",
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
			cfg: config.SeqAPI{
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
				},
				DefaultEnv: "test",
			},
			want: &seqapi.GetEnvsResponse{
				Envs: []*seqapi.GetEnvsResponse_Env{
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
				envsResponse: parseEnvs(tt.cfg),
			}
			resp, err := api.GetEnvs(context.TODO(), &seqapi.GetEnvsRequest{})
			require.NoError(t, err)
			require.True(t, proto.Equal(tt.want, resp))
		})
	}
}
