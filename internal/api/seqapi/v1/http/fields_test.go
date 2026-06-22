package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetFields(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.GetFieldsResponse
		err  error
	}

	tests := []struct {
		name string

		cfg config.SeqAPIOptions

		want    getFieldsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: getFieldsResponse{
				Fields: fields{
					{Name: "test_name1", Type: "keyword"},
					{Name: "test_name2", Type: "text"},
				},
			},
			mockArgs: &mockArgs{
				resp: &seqapi.GetFieldsResponse{
					Fields: []*seqapi.Field{
						{
							Name: "test_name1",
							Type: seqapi.FieldType_keyword,
						},
						{
							Name: "test_name2",
							Type: seqapi.FieldType_text,
						},
					},
				},
			},
		},
		{
			name: "ok_with_system_and_pinned_fields",
			want: getFieldsResponse{
				Fields: fields{
					{Name: "test_name1", Type: "keyword"},
					{Name: "test_name2", Type: "text"},
				},
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
					Fields: []*seqapi.Field{
						{
							Name: "test_name1",
							Type: seqapi.FieldType_keyword,
						},
						{
							Name: "test_name2",
							Type: seqapi.FieldType_text,
						},
					},
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
			seqDbMock.EXPECT().
				GetFields(gomock.Any(), gomock.Any()).
				Return(tt.mockArgs.resp, tt.mockArgs.err).
				Times(1)
			seqData.Mocks.SeqDB = seqDbMock

			api := setupAPI(seqData)

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

func TestServeGetFieldsCached(t *testing.T) {
	tests := []struct {
		name string

		resp *seqapi.GetFieldsResponse
		want getFieldsResponse
	}{
		{
			name: "ok",
			resp: &seqapi.GetFieldsResponse{
				Fields: []*seqapi.Field{
					{
						Name: "n1",
						Type: seqapi.FieldType_keyword,
					},
					{
						Name: "n2",
						Type: seqapi.FieldType_text,
					},
				},
			},
			want: getFieldsResponse{
				Fields: fields{
					{Name: "n1", Type: "keyword"},
					{Name: "n2", Type: "text"},
				},
			},
		},
		{
			name: "another_ok",
			resp: &seqapi.GetFieldsResponse{
				Fields: []*seqapi.Field{
					{
						Name: "qwe",
						Type: seqapi.FieldType_keyword,
					},
				},
			},
			want: getFieldsResponse{
				Fields: fields{
					{Name: "qwe", Type: "keyword"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seqDbMock := mock_seqdb.NewMockClient(ctrl)

			seqDbMock.EXPECT().
				GetFields(gomock.Any(), gomock.Any()).
				Return(tt.resp, nil).
				Times(1)

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						FieldsCacheTTL: ttl,
					},
				},
				Mocks: test.Mocks{
					SeqDB: seqDbMock,
				},
			}

			api := setupAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFieldsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/fields",
				Handler: api.serveGetFields,
				Want:    tt.want,
			})

			time.Sleep(ttl / 2)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFieldsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/fields",
				Handler: api.serveGetFields,
				Want:    tt.want,
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

			api := setupAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFieldsResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/fields/pinned",
				Handler: api.serveGetPinnedFields,
				Want:    tt.want,
			})
		})
	}
}
