package http

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetFields(t *testing.T) {
	f := fields{
		{Name: "test_name1", Type: "keyword"},
		{Name: "test_name2", Type: "text"},
	}
	fAPI := []*seqapi.Field{
		{Name: "test_name1", Type: seqapi.FieldType_keyword},
		{Name: "test_name2", Type: seqapi.FieldType_text},
	}

	marshal := func(data any) []byte {
		raw, err := json.Marshal(data)
		require.NoError(t, err)
		return raw
	}

	type mockArgs struct {
		resp *seqapi.GetFieldsResponse
		err  error
	}

	tests := []struct {
		name string

		cfg       config.SeqAPIOptions
		cacheData []byte

		want    getFieldsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: getFieldsResponse{
				Fields: f,
			},
			mockArgs: &mockArgs{
				resp: &seqapi.GetFieldsResponse{
					Fields: fAPI,
				},
			},
		},
		{
			name: "ok_with_system_and_pinned_fields",
			want: getFieldsResponse{
				Fields: f,
				SystemFields: fields{
					{Name: "field1", Type: "keyword"},
					{Name: "field2", Type: "text"},
				},
				PinnedFields: fields{
					{Name: "field3", Type: "keyword"},
					{Name: "field4", Type: "text"},
				},
			},
			mockArgs: &mockArgs{
				resp: &seqapi.GetFieldsResponse{
					Fields: fAPI,
				},
			},
			cfg: config.SeqAPIOptions{
				SystemFields: []config.Field{
					{Name: "field1", Type: "keyword"},
					{Name: "field2", Type: "text"},
				},
				PinnedFields: []config.Field{
					{Name: "field3", Type: "keyword"},
					{Name: "field4", Type: "text"},
				},
			},
		},
		{
			name: "ok_cached",
			want: getFieldsResponse{
				Fields: f,
			},
			cacheData: marshal(getFieldsResponse{Fields: f}),
			cfg: config.SeqAPIOptions{
				FieldsCacheTTL: time.Hour,
			},
		},
		{
			name:    "err_client",
			wantErr: true,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &tt.cfg,
				},
			}

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqData.Mocks.SeqDB = seqDbMock

			api := setupTestAPI(seqData)

			if tt.mockArgs != nil {
				seqDbMock.EXPECT().
					GetFields(gomock.Any(), gomock.Any()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}
			if tt.cacheData != nil && tt.cfg.FieldsCacheTTL > 0 {
				api.params.fieldsCache.setFields(tt.cacheData)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFieldsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/fields",
				Handler: api.serveGetFields,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeGetPinnedFields(t *testing.T) {
	tests := []struct {
		name string

		fields []config.Field
		want   getFieldsResponse
	}{
		{
			name: "ok",
			fields: []config.Field{
				{Name: "field1", Type: "keyword"},
				{Name: "field2", Type: "text"},
			},
			want: getFieldsResponse{
				Fields: fields{
					{Name: "field1", Type: "keyword"},
					{Name: "field2", Type: "text"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						PinnedFields: tt.fields,
					},
				},
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFieldsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/fields/pinned",
				Handler: api.serveGetPinnedFields,
				Want:    tt.want,
			})
		})
	}
}
