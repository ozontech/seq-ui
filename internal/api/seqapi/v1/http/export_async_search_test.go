package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeExportAsyncSearch(t *testing.T) {
	type mockArgs struct {
		req *seqapi.ExportAsyncSearchRequest
		err error
	}

	tests := []struct {
		name string

		req     exportAsyncSearchRequest
		cfg     config.SeqAPI
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_jsonl",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    50,
				Offset:   0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.ExportAsyncSearchRequest{
					SearchId: testSearchID,
					Limit:    50,
					Offset:   0,
					Format:   seqapi.ExportFormat_EXPORT_FORMAT_JSONL,
				},
			},
		},
		{
			name: "ok_csv",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    50,
				Offset:   0,
				Format:   efCSV,
				Fields:   []string{"field1", "field2"},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.ExportAsyncSearchRequest{
					SearchId: testSearchID,
					Limit:    50,
					Offset:   0,
					Format:   seqapi.ExportFormat_EXPORT_FORMAT_CSV,
					Fields:   []string{"field1", "field2"},
				},
			},
		},
		{
			name: "err_parallel_limited",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    0,
				Offset:   0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxParallelExportRequests: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "err_export_limit_max",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    10,
				Offset:   0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            5,
					MaxParallelExportRequests: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "err_csv_empty_fields",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    10,
				Offset:   0,
				Format:   efCSV,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "err_invalid_id",
			req: exportAsyncSearchRequest{
				SearchID: "some invalid id",
				Limit:    10,
				Offset:   0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "err_client",
			req: exportAsyncSearchRequest{
				SearchID: testSearchID,
				Limit:    50,
				Offset:   0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: &seqapi.ExportAsyncSearchRequest{
					SearchId: testSearchID,
					Limit:    50,
					Offset:   0,
					Format:   seqapi.ExportFormat_EXPORT_FORMAT_JSONL,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			svcMock := mock_asyncsearches.NewMockService(ctrl)

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}
			seqData.Mocks.AsyncSearchesSvc = svcMock

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					ExportAsyncSearch(gomock.Any(), tt.mockArgs.req, gomock.Any()).
					Return(tt.mockArgs.err).
					Times(1)
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[exportAsyncSearchRequest, struct{}]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/async_search/export",
				Req:     tt.req,
				Handler: api.serveExportAsyncSearch,
				WantErr: tt.wantErr,
				NoResp:  true,
			})
		})
	}
}

func TestServeExportAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[exportAsyncSearchRequest, struct{}]{
		Method:  http.MethodPost,
		Target:  "/seqapi/v1/async_search/export",
		Handler: api.serveExportAsyncSearch,
		WantErr: true,
	})
}
