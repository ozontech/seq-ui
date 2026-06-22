package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeExport(t *testing.T) {
	type mockArgs struct {
		req *seqapi.ExportRequest
		err error
	}

	tests := []struct {
		name string

		req     exportRequest
		cfg     config.SeqAPI
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_jsonl",
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  50,
				Offset: 0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.ExportRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  50,
					Offset: 0,
				},
			},
		},
		{
			name: "ok_csv",
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  50,
				Offset: 0,
				Format: efCSV,
				Fields: []string{"field1", "field2"},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.ExportRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  50,
					Offset: 0,
					Format: seqapi.ExportFormat_EXPORT_FORMAT_CSV,
					Fields: []string{"field1", "field2"},
				},
			},
		},
		{
			name: "err_parallel_limited",
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  0,
				Offset: 0,
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
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  10,
				Offset: 0,
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
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  10,
				Offset: 0,
				Format: efCSV,
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
			req: exportRequest{
				Query:  query,
				From:   from,
				To:     to,
				Limit:  50,
				Offset: 0,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
			wantErr: true,
			mockArgs: &mockArgs{
				req: &seqapi.ExportRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  50,
					Offset: 0,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)
				seqDbMock := mock_seqdb.NewMockClient(ctrl)

				seqDbMock.EXPECT().
					Export(gomock.Any(), tt.mockArgs.req, gomock.Any()).
					Return(tt.mockArgs.err).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[exportRequest, struct{}]{
				Method:  http.MethodPost,
				Target:  "/seqapi/v1/export",
				Req:     tt.req,
				Handler: api.serveExport,
				WantErr: tt.wantErr,
				NoResp:  true,
			})
		})
	}
}
