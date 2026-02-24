package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/stretchr/testify/require"
)

func TestServeGetEnvs(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.SeqAPI
		wantEnvs []envInfo
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
			wantEnvs: []envInfo{
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
		{
			name: "ok_multiple_envs",
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{},
				Envs: map[string]config.SeqAPIEnv{
					"prod": {
						SeqDB: "prod-seqdb",
						Options: &config.SeqAPIOptions{
							MaxSearchLimit:            1000,
							MaxExportLimit:            500,
							MaxParallelExportRequests: 10,
							MaxAggregationsPerRequest: 5,
							SeqCLIMaxSearchLimit:      2000,
						},
					},
					"staging": {
						SeqDB: "staging-seqdb",
						Options: &config.SeqAPIOptions{
							MaxSearchLimit:            500,
							MaxExportLimit:            250,
							MaxParallelExportRequests: 5,
							MaxAggregationsPerRequest: 3,
							SeqCLIMaxSearchLimit:      1000,
						},
					},
				},
				DefaultEnv: "prod",
			},
			wantEnvs: []envInfo{
				{
					Env:                       "prod",
					MaxSearchLimit:            1000,
					MaxExportLimit:            500,
					MaxParallelExportRequests: 10,
					MaxAggregationsPerRequest: 5,
					SeqCliMaxSearchLimit:      2000,
				},
				{
					Env:                       "staging",
					MaxSearchLimit:            500,
					MaxExportLimit:            250,
					MaxParallelExportRequests: 5,
					MaxAggregationsPerRequest: 3,
					SeqCliMaxSearchLimit:      1000,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			api := initTestAPI(seqData)

			req := httptest.NewRequest(http.MethodGet, "/seqapi/v1/envs", http.NoBody)
			w := httptest.NewRecorder()
			api.serveGetEnvs(w, req)

			var response getEnvsResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err, "failed to decode response")
			require.ElementsMatch(t, tt.wantEnvs, response.Envs, "Returned envs do not match expected")
		})
	}
}
