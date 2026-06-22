package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestGetFields(t *testing.T) {
	tests := []struct {
		name string

		cfg       config.SeqAPIOptions
		seqDBResp *seqapi.GetFieldsResponse
		wantResp  *seqapi.GetFieldsResponse

		clientErr error
	}{
		{
			name: "ok",
			seqDBResp: &seqapi.GetFieldsResponse{
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
			wantResp: &seqapi.GetFieldsResponse{
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
		{
			name: "ok_with_system_and_pinned_fields",
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
			seqDBResp: &seqapi.GetFieldsResponse{
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
			wantResp: &seqapi.GetFieldsResponse{
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
				SystemFields: []*seqapi.Field{
					{Name: "field1", Type: seqapi.FieldType_keyword},
					{Name: "field2", Type: seqapi.FieldType_text},
				},
				PinnedFields: []*seqapi.Field{
					{Name: "field3", Type: seqapi.FieldType_keyword},
					{Name: "field4", Type: seqapi.FieldType_text},
				},
			},
		},
		{
			name:      "err_client",
			clientErr: errors.New("client error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seqDbMock := mock_seqdb.NewMockClient(ctrl)

			seqDbMock.EXPECT().
				GetFields(gomock.Any(), nil).
				Return(proto.Clone(tt.seqDBResp), tt.clientErr).
				Times(1)

			seqData := test.APITestData{
				Mocks: test.Mocks{
					SeqDB: seqDbMock,
				},
				Cfg: config.SeqAPI{
					SeqAPIOptions: &tt.cfg,
				},
			}

			api := setupAPI(seqData)
			resp, err := api.GetFields(context.Background(), nil)

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(tt.wantResp, resp))
		})
	}
}

func TestGetFieldsCached(t *testing.T) {
	responses := []*seqapi.GetFieldsResponse{
		{
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
		{
			Fields: []*seqapi.Field{
				{
					Name: "qwe",
					Type: seqapi.FieldType_text,
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	seqDbMock := mock_seqdb.NewMockClient(ctrl)

	for _, r := range responses {
		seqDbMock.EXPECT().
			GetFields(gomock.Any(), nil).
			Return(proto.Clone(r), nil).
			Times(1)
	}

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

	for _, r := range responses {
		resp, err := api.GetFields(context.Background(), nil)
		require.NoError(t, err)
		require.True(t, proto.Equal(r, resp))

		time.Sleep(ttl / 2)

		resp, err = api.GetFields(context.Background(), nil)
		require.NoError(t, err)
		require.True(t, proto.Equal(r, resp))

		time.Sleep(ttl)
	}
}

func TestGetPinnedFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []config.Field
	}{
		{
			name: "ok",
			fields: []config.Field{
				{Name: "field1", Type: "keyword"},
				{Name: "field2", Type: "text"},
			},
		},
		{
			name: "empty",
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

			resp, err := api.GetPinnedFields(context.Background(), nil)
			require.NoError(t, err)

			require.Equal(t, len(tt.fields), len(resp.Fields))
			for i, f := range resp.Fields {
				require.Equal(t, tt.fields[i].Name, f.Name)
				require.Equal(t, tt.fields[i].Type, f.Type.String())
			}
		})
	}
}
