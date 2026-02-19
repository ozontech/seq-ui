package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStatus(t *testing.T) {
	type TestCase struct {
		name      string
		resp      *seqapi.StatusResponse
		clientErr error
	}

	someMoment := time.Now()

	tests := []TestCase{
		{
			name: "ok",
			resp: &seqapi.StatusResponse{
				NumberOfStores:    3,
				OldestStorageTime: timestamppb.New(someMoment),
				Stores: []*seqapi.StoreStatus{
					{
						Host:   "host-0",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(someMoment)},
					},
					{
						Host:   "host-1",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(someMoment.Add(1 * time.Hour))},
					},
					{
						Host:   "host-2",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(someMoment.Add(2 * time.Hour))},
					},
				},
			},
		},
		{
			name:      "err_client",
			clientErr: errors.New("client error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			seqDbMock.EXPECT().Status(gomock.Any(), nil).
				Return(proto.Clone(tt.resp), tt.clientErr).Times(1)

			cfg := config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB:   "test",
						Options: &config.SeqAPIOptions{},
					},
				},
				DefaultEnv: "test",
			}

			seqData := test.APITestData{
				Cfg: cfg,
				Mocks: test.Mocks{
					SeqDB: seqDbMock,
				},
			}
			s := initTestAPI(seqData)

			md := metadata.New(map[string]string{"env": "test"})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := s.Status(ctx, nil)

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}
