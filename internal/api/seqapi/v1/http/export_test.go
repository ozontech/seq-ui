package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServeExport(t *testing.T) {
	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)

	formatReqBody := func(limit int, format exportFormat, fields []string) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`{"query":%q,"from":%q,"to":%q,"offset":0,"limit":%d`,
			query, from.Format(time.RFC3339), to.Format(time.RFC3339), limit))

		if format != "" {
			sb.WriteString(fmt.Sprintf(`,"format":%q`, format))
		}
		if len(fields) > 0 {
			fieldsRaw, err := json.Marshal(fields)
			assert.NoError(t, err)
			sb.WriteString(fmt.Sprintf(`,"fields":%s`, fieldsRaw))
		}

		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req *seqapi.ExportRequest
		err error
	}

	tests := []struct {
		name string

		reqBody    string
		wantStatus int

		mockArgs *mockArgs
		cfg      config.SeqAPI
	}{
		{
			name:    "ok_jsonl",
			reqBody: formatReqBody(50, "", nil),
			mockArgs: &mockArgs{
				req: &seqapi.ExportRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  50,
					Offset: 0,
				},
			},
			wantStatus: http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
		},
		{
			name:    "ok_csv",
			reqBody: formatReqBody(50, efCSV, []string{"field1", "field2"}),
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
			wantStatus: http.StatusOK,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_parallel_limited",
			reqBody:    formatReqBody(0, "", nil),
			wantStatus: http.StatusTooManyRequests,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxParallelExportRequests: 0,
				},
			},
		},
		{
			name:       "err_export_limit_max",
			reqBody:    formatReqBody(10, "", nil),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            5,
					MaxParallelExportRequests: 1,
				},
			},
		},
		{
			name:       "err_csv_empty_fields",
			reqBody:    formatReqBody(10, efCSV, nil),
			wantStatus: http.StatusBadRequest,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
				},
			},
		},
		{
			name:    "err_client",
			reqBody: formatReqBody(50, "", nil),
			mockArgs: &mockArgs{
				req: &seqapi.ExportRequest{
					Query:  query,
					From:   timestamppb.New(from),
					To:     timestamppb.New(to),
					Limit:  50,
					Offset: 0,
				},
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxExportLimit:            100,
					MaxParallelExportRequests: 1,
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

			req := httptest.NewRequest(http.MethodPost, "/seqapi/v1/export", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)
				seqDbMock := mock_seqdb.NewMockClient(ctrl)

				cw, _ := httputil.NewChunkedWriter(w)
				seqDbMock.EXPECT().Export(gomock.Any(), tt.mockArgs.req, cw).
					Return(tt.mockArgs.err).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			s := initTestAPI(seqData)

			s.serveExport(w, req)

			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantStatus, res.StatusCode)
		})
	}
}
