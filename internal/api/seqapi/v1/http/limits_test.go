package http

import (
	"net/http"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
)

func TestServeGetLimits(t *testing.T) {
	tests := []struct {
		name string

		env  string
		cfg  config.SeqAPI
		want getLimitsResponse
	}{
		{
			name: "ok",
			env:  "default",
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:            100,
					MaxExportLimit:            200,
					MaxParallelExportRequests: 2,
					MaxAggregationsPerRequest: 5,
					SeqCLIMaxSearchLimit:      10000,
				},
			},
			want: getLimitsResponse{
				MaxSearchLimit:            100,
				MaxExportLimit:            200,
				MaxParallelExportRequests: 2,
				MaxAggregationsPerRequest: 5,
				SeqCliMaxSearchLimit:      10000,
			},
		},
		{
			name: "empty",
			env:  "default",
			want: getLimitsResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getLimitsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/limits",
				Handler: api.serveGetLimits,
				Want:    tt.want,
			})
		})
	}
}
