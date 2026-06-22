package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestGetHistogram(t *testing.T) {
	var (
		query    = "message:error"
		interval = "2s"
	)
	tests := []struct {
		name string

		req  *seqapi.GetHistogramRequest
		want *seqapi.GetHistogramResponse

		clientErr error
	}{
		{
			name: "ok",
			req: &seqapi.GetHistogramRequest{
				Query:    query,
				From:     timestamppb.New(testFrom),
				To:       timestamppb.New(testTo),
				Interval: interval,
			},
			want: &seqapi.GetHistogramResponse{
				Histogram: test.MakeHistogram(5),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
		},
		{
			name: "err_client",
			req: &seqapi.GetHistogramRequest{
				Interval: interval,
			},
			clientErr: errSomethingWrong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			ctrl := gomock.NewController(t)

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqDbMock.EXPECT().
				GetHistogram(gomock.Any(), proto.Clone(tt.req)).
				Return(proto.Clone(tt.want), tt.clientErr).
				Times(1)
			seqData.Mocks.SeqDB = seqDbMock

			api := setupTestAPI(seqData)
			resp, err := api.GetHistogram(context.Background(), tt.req)

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(tt.want, resp))
		})
	}
}
