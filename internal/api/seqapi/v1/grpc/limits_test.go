package grpc

import (
	"context"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestGetLimits(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.SeqAPI
		want *seqapi.GetLimitsResponse
	}{
		{
			name: "ok",
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB: "test",
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
			want: &seqapi.GetLimitsResponse{
				MaxSearchLimit:            100,
				MaxExportLimit:            200,
				MaxParallelExportRequests: 2,
				MaxAggregationsPerRequest: 5,
				SeqCliMaxSearchLimit:      10000,
			},
		},
		{
			name: "empty",
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB:   "test",
						Options: &config.SeqAPIOptions{},
					},
				},
				DefaultEnv: "test",
			},
			want: &seqapi.GetLimitsResponse{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}
			s := initTestAPI(seqData)

			md := metadata.New(map[string]string{"env": "test"})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := s.GetLimits(ctx, nil)

			require.NoError(t, err)
			require.True(t, proto.Equal(tt.want, resp))
		})
	}
}
