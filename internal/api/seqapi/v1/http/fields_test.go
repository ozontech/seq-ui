package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
)

func TestServeGetFields(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.GetFieldsResponse
		err  error
	}

	tests := []struct {
		name string

		wantRespBody string
		wantStatus   int

		mockArgs mockArgs
	}{
		{
			name: "ok",
			mockArgs: mockArgs{
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
			wantRespBody: `{"fields":[{"name":"test_name1","type":"keyword"},{"name":"test_name2","type":"text"}]}`,
			wantStatus:   http.StatusOK,
		},
		{
			name: "err_client",
			mockArgs: mockArgs{
				err: errors.New("client error"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			seqData := test.APITestData{}

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqDbMock.EXPECT().GetFields(gomock.Any(), gomock.Any()).
				Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			seqData.Mocks.SeqDB = seqDbMock

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodGet, "/seqapi/v1/fields", http.NoBody)

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetFields,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeGetFieldsCached(t *testing.T) {
	type TestCase struct {
		resp         *seqapi.GetFieldsResponse
		wantRespBody string
	}

	tests := []TestCase{
		{
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
			wantRespBody: `{"fields":[{"name":"n1","type":"keyword"},{"name":"n2","type":"text"}]}`,
		},
		{
			resp: &seqapi.GetFieldsResponse{
				Fields: []*seqapi.Field{
					{
						Name: "qwe",
						Type: seqapi.FieldType_keyword,
					},
				},
			},
			wantRespBody: `{"fields":[{"name":"qwe","type":"keyword"}]}`,
		},
	}

	ctrl := gomock.NewController(t)
	seqDbMock := mock_seqdb.NewMockClient(ctrl)

	for _, testCase := range tests {
		seqDbMock.EXPECT().GetFields(gomock.Any(), gomock.Any()).
			Return(testCase.resp, nil).Times(1)
	}

	const ttl = 20 * time.Millisecond

	seqData := test.APITestData{
		Cfg: config.SeqAPI{
			FieldsCacheTTL: ttl,
		},
		Mocks: test.Mocks{
			SeqDB: seqDbMock,
		},
	}

	api := initTestAPI(seqData)

	for _, testCase := range tests {
		httputil.DoTestHTTP(t, httputil.TestDataHTTP{
			Req:          httptest.NewRequest(http.MethodGet, "/seqapi/v1/fields", http.NoBody),
			Handler:      api.serveGetFields,
			WantRespBody: testCase.wantRespBody,
			WantStatus:   http.StatusOK,
		})

		time.Sleep(ttl / 2)

		httputil.DoTestHTTP(t, httputil.TestDataHTTP{
			Req:          httptest.NewRequest(http.MethodGet, "/seqapi/v1/fields", http.NoBody),
			Handler:      api.serveGetFields,
			WantRespBody: testCase.wantRespBody,
			WantStatus:   http.StatusOK,
		})

		time.Sleep(ttl)
	}
}

func TestServeGetPinnedFields(t *testing.T) {
	tests := []struct {
		name string

		fields       []config.PinnedField
		wantRespBody string
	}{
		{
			name: "ok",
			fields: []config.PinnedField{
				{Name: "field1", Type: "keyword"},
				{Name: "field2", Type: "text"},
			},
			wantRespBody: `{"fields":[{"name":"field1","type":"keyword"},{"name":"field2","type":"text"}]}`,
		},
		{
			name:         "empty",
			wantRespBody: `{"fields":[]}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					PinnedFields: tt.fields,
				},
			}

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodGet, "/seqapi/v1/fields/pinned", http.NoBody)
			w := httptest.NewRecorder()

			api.serveGetPinnedFields(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			var gfr getFieldsResponse
			err = json.Unmarshal(respBody, &gfr)
			assert.NoError(t, err)

			require.Equal(t, len(tt.fields), len(gfr.Fields))
			for i, f := range gfr.Fields {
				require.Equal(t, tt.fields[i].Name, f.Name)
				require.Equal(t, tt.fields[i].Type, f.Type)
			}
		})
	}
}
