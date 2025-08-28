package seqdb

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1"
	mock "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/seqproxyapi/v1/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_GRPCClient_GetFields(t *testing.T) {
	type mockArgs struct {
		resp *seqproxyapi.MappingResponse
		err  error
	}

	prepareMockArgs := func(fields map[string]string, err error) mockArgs {
		var proxyResp *seqproxyapi.MappingResponse

		if len(fields) != 0 {
			data, err := json.Marshal(fields)
			assert.NoError(t, err)
			proxyResp = &seqproxyapi.MappingResponse{
				Data: data,
			}
		}

		return mockArgs{
			resp: proxyResp,
			err:  err,
		}
	}

	tests := []struct {
		name string

		wantFields map[string]string
		wantErr    error
	}{
		{
			name: "ok",
			wantFields: map[string]string{
				"test_name1": seqapi.FieldType_keyword.String(),
				"test_name2": seqapi.FieldType_text.String(),
			},
		},
		{
			name:    "err_proxy",
			wantErr: errors.New("proxy error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mArgs := prepareMockArgs(tt.wantFields, tt.wantErr)

			ctrl := gomock.NewController(t)
			seqProxyMock := mock.NewMockSeqProxyApiClient(ctrl)
			seqProxyMock.EXPECT().Mapping(ctx, &seqproxyapi.MappingRequest{}).
				Return(mArgs.resp, mArgs.err).Times(1)

			c := initGRPCClient(seqProxyMock)

			resp, err := c.GetFields(ctx, &seqapi.GetFieldsRequest{})

			require.Equal(t, tt.wantErr, err)
			if len(tt.wantFields) != 0 {
				require.Equal(t, len(tt.wantFields), len(resp.Fields))
				for _, f := range resp.Fields {
					require.Equal(t, tt.wantFields[f.Name], f.Type.String())
				}
			}
		})
	}
}
