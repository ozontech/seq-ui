package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetHistogram(t *testing.T) {
	query := "message:error"
	from := time.Now()
	to := from.Add(time.Second)
	interval := "5s"

	tests := []struct {
		name string

		req  *seqapi.GetHistogramRequest
		resp *seqapi.GetHistogramResponse

		clientErr error
	}{
		{
			name: "ok",
			req: &seqapi.GetHistogramRequest{
				Query:    query,
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				Interval: interval,
			},
			resp: &seqapi.GetHistogramResponse{
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
			clientErr: errors.New("client error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			ctrl := gomock.NewController(t)

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqDbMock.EXPECT().GetHistogram(gomock.Any(), proto.Clone(tt.req)).
				Return(proto.Clone(tt.resp), tt.clientErr).Times(1)

			seqData.Mocks.SeqDB = seqDbMock

			s := initTestAPI(seqData)

			resp, err := s.GetHistogram(context.Background(), tt.req)

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}
