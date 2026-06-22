package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestStatus(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.StatusResponse
		err  error
	}

	tests := []struct {
		name string

		want     *seqapi.StatusResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: &seqapi.StatusResponse{
				NumberOfStores:    3,
				OldestStorageTime: timestamppb.New(testSomeMoment),
				Stores: []*seqapi.StoreStatus{
					{
						Host:   "host-0",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment)},
					},
					{
						Host:   "host-1",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment.Add(1 * time.Hour))},
					},
					{
						Host:   "host-2",
						Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment.Add(2 * time.Hour))},
					},
				},
			},
			mockArgs: &mockArgs{
				resp: &seqapi.StatusResponse{
					NumberOfStores:    3,
					OldestStorageTime: timestamppb.New(testSomeMoment),
					Stores: []*seqapi.StoreStatus{
						{
							Host:   "host-0",
							Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment)},
						},
						{
							Host:   "host-1",
							Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment.Add(1 * time.Hour))},
						},
						{
							Host:   "host-2",
							Values: &seqapi.StoreStatusValues{OldestTime: timestamppb.New(testSomeMoment.Add(2 * time.Hour))},
						},
					},
				},
			},
		},
		{
			name:     "err_client",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				err: status.Error(codes.Internal, "client error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			seqDbMock := mock_seqdb.NewMockClient(ctrl)

			seqDbMock.EXPECT().
				Status(gomock.Any(), nil).
				Return(proto.Clone(tt.mockArgs.resp), tt.mockArgs.err).
				Times(1)

			seqData := test.APITestData{
				Mocks: test.Mocks{
					SeqDB: seqDbMock,
				},
			}

			api := setupTestAPI(seqData)

			got, err := api.Status(context.Background(), nil)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.True(t, proto.Equal(tt.want, got))
		})
	}
}
