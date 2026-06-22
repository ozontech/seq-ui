package grpc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestGetLogsLifespan(t *testing.T) {
	var (
		resultStr = "36000" // 10(h) * 60(min/h) * 60(sec/min)
		cacheKey  = "logs_lifespan"
		result    = 10 * time.Hour
		cacheTTL  = time.Minute
	)
	unparsable := func(s string) bool {
		_, err := strconv.Atoi(s)
		return err != nil
	}

	tests := []struct {
		name string

		getOp test.CacheMockArgs
		setOp test.CacheMockArgs

		clientResp *seqapi.StatusResponse
		clientErr  error

		resp *seqapi.GetLogsLifespanResponse
	}{
		{
			name: "ok_cached",
			getOp: test.CacheMockArgs{
				Value: resultStr,
			},
			resp: &seqapi.GetLogsLifespanResponse{
				Lifespan: durationpb.New(result),
			},
		},
		{
			name: "ok_cached_unparsable",
			getOp: test.CacheMockArgs{
				Value: "10h", // value format changed
			},
			setOp: test.CacheMockArgs{
				Value: resultStr,
			},
			clientResp: &seqapi.StatusResponse{
				OldestStorageTime: timestamppb.New(testSomeMoment),
			},
			resp: &seqapi.GetLogsLifespanResponse{
				Lifespan: durationpb.New(result),
			},
		},
		{
			name: "ok_no_cached",
			getOp: test.CacheMockArgs{
				Err: cache.ErrNotFound,
			},
			setOp: test.CacheMockArgs{
				Value: resultStr,
			},
			clientResp: &seqapi.StatusResponse{
				OldestStorageTime: timestamppb.New(testSomeMoment),
			},
			resp: &seqapi.GetLogsLifespanResponse{
				Lifespan: durationpb.New(result),
			},
		},
		{
			name: "err_client",
			getOp: test.CacheMockArgs{
				Err: cache.ErrNotFound,
			},
			clientErr: errSomethingWrong,
		},
		{
			name: "err_nil_oldest_storage_time",
			getOp: test.CacheMockArgs{
				Err: cache.ErrNotFound,
			},
			clientResp: &seqapi.StatusResponse{
				OldestStorageTime: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						LogsLifespanCacheKey: cacheKey,
						LogsLifespanCacheTTL: cacheTTL,
					},
				},
			}

			ctrl := gomock.NewController(t)
			cacheMock := mock_cache.NewMockCache(ctrl)

			cacheMock.EXPECT().
				Get(gomock.Any(), cacheKey).
				Return(tt.getOp.Value, tt.getOp.Err).
				Times(1)
			seqData.Mocks.Cache = cacheMock

			if tt.getOp.Err != nil || unparsable(tt.getOp.Value) {
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().
					Status(gomock.Any(), gomock.Any()).
					Return(proto.Clone(tt.clientResp), tt.clientErr).
					Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				if tt.clientErr == nil && tt.clientResp.OldestStorageTime != nil {
					cacheMock.EXPECT().
						SetWithTTL(gomock.Any(), cacheKey, tt.setOp.Value, cacheTTL).
						Return(tt.setOp.Err).
						Times(1)
				}
			}

			s := setupTestAPI(seqData)
			s.nowFn = func() time.Time {
				return testSomeMoment.Add(result)
			}

			resp, err := s.GetLogsLifespan(context.Background(), nil)
			if tt.resp == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}
