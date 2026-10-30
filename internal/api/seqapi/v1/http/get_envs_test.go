package http

import (
	"net/http"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
)

func TestServeGetEnvs(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.SeqAPI
		want getEnvsResponse
	}{
		{
			name: "single_env",
			cfg: config.SeqAPI{
				Options: config.SeqAPIOptions{
					Limits: config.SeqAPILimits{
						MaxSearchLimit:            100,
						MaxExportLimit:            200,
						MaxParallelExportRequests: 2,
						MaxAggregationsPerRequest: 5,
						SeqCLIMaxSearchLimit:      10000,
					},
				},
			},
			want: getEnvsResponse{
				[]envInfo{
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
			name: "ok_multiple_envs",
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"cluster-220": {
						Options: &config.SeqAPIOptions{
							SeqDBID: "pro-seqdb",
							Limits: config.SeqAPILimits{
								MaxSearchLimit:            1000,
								MaxExportLimit:            500,
								MaxParallelExportRequests: 10,
								MaxAggregationsPerRequest: 5,
								SeqCLIMaxSearchLimit:      2000,
							},
						},
					},
					"cluster-10": {
						Options: &config.SeqAPIOptions{
							SeqDBID: "prod-seqdb",
							Limits: config.SeqAPILimits{
								MaxSearchLimit:            1000,
								MaxExportLimit:            500,
								MaxParallelExportRequests: 10,
								MaxAggregationsPerRequest: 5,
								SeqCLIMaxSearchLimit:      2000,
							},
						},
					},
					"cluster-102": {
						Options: &config.SeqAPIOptions{
							SeqDBID: "staging-seqdb",
							Limits: config.SeqAPILimits{
								MaxSearchLimit:            500,
								MaxExportLimit:            250,
								MaxParallelExportRequests: 5,
								MaxAggregationsPerRequest: 3,
								SeqCLIMaxSearchLimit:      1000,
							},
						},
					},
					"prod": {
						Options: &config.SeqAPIOptions{
							SeqDBID: "stag-seqdb",
							Limits: config.SeqAPILimits{
								MaxSearchLimit:            500,
								MaxExportLimit:            250,
								MaxParallelExportRequests: 5,
								MaxAggregationsPerRequest: 3,
								SeqCLIMaxSearchLimit:      1000,
							},
						},
					},
					"wyanki": {
						Options: &config.SeqAPIOptions{
							SeqDBID: "sta-seqdb",
							Limits: config.SeqAPILimits{
								MaxSearchLimit:            500,
								MaxExportLimit:            250,
								MaxParallelExportRequests: 5,
								MaxAggregationsPerRequest: 3,
								SeqCLIMaxSearchLimit:      1000,
							},
						},
					},
				},
				DefaultEnv: "cluster-10",
			},
			want: getEnvsResponse{
				[]envInfo{
					{
						Env:                       "cluster-10",
						MaxSearchLimit:            1000,
						MaxExportLimit:            500,
						MaxParallelExportRequests: 10,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      2000,
					},
					{
						Env:                       "cluster-102",
						MaxSearchLimit:            500,
						MaxExportLimit:            250,
						MaxParallelExportRequests: 5,
						MaxAggregationsPerRequest: 3,
						SeqCliMaxSearchLimit:      1000,
					},
					{
						Env:                       "cluster-220",
						MaxSearchLimit:            1000,
						MaxExportLimit:            500,
						MaxParallelExportRequests: 10,
						MaxAggregationsPerRequest: 5,
						SeqCliMaxSearchLimit:      2000,
					},
					{
						Env:                       "prod",
						MaxSearchLimit:            500,
						MaxExportLimit:            250,
						MaxParallelExportRequests: 5,
						MaxAggregationsPerRequest: 3,
						SeqCliMaxSearchLimit:      1000,
					},
					{
						Env:                       "wyanki",
						MaxSearchLimit:            500,
						MaxExportLimit:            250,
						MaxParallelExportRequests: 5,
						MaxAggregationsPerRequest: 3,
						SeqCliMaxSearchLimit:      1000,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getEnvsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/envs",
				Handler: api.serveGetEnvs,
				Want:    tt.want,
			})
		})
	}
}
