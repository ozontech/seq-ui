package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestServeGetLimits(t *testing.T) {
	tests := []struct {
		name         string
		env          string
		cfg          config.SeqAPI
		wantRespBody string
	}{
		{
			name: "ok",
			env:  "prod",
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"prod": {
						SeqDB: "prod",
						Options: &config.SeqAPIOptions{
							MaxSearchLimit:            100,
							MaxExportLimit:            200,
							MaxParallelExportRequests: 2,
							MaxAggregationsPerRequest: 5,
							SeqCLIMaxSearchLimit:      10000,
						},
					},
				},
			},
			wantRespBody: `{"maxSearchLimit":100,"maxExportLimit":200,"maxParallelExportRequests":2,"maxAggregationsPerRequest":5,"seqCliMaxSearchLimit":10000}`,
		},
		{
			name:         "empty",
			env:          "prod",
			wantRespBody: `{"maxSearchLimit":0,"maxExportLimit":0,"maxParallelExportRequests":0,"maxAggregationsPerRequest":0,"seqCliMaxSearchLimit":0}`,
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
			req := httptest.NewRequest(http.MethodGet, "/seqapi/v1/limits", http.NoBody)

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetLimits,
				WantRespBody: tt.wantRespBody,
				WantStatus:   http.StatusOK,
			})
		})
	}
}
