package http

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetLogsLifespan(t *testing.T) {
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

		wantErr bool
		want    getLogsLifespanResponse
	}{
		{
			name: "ok_cached",
			getOp: test.CacheMockArgs{
				Value: resultStr,
			},
			want: getLogsLifespanResponse{Lifespan: 36000},
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
			want: getLogsLifespanResponse{Lifespan: 36000},
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
			want: getLogsLifespanResponse{Lifespan: 36000},
		},
		{
			name: "err_client",
			getOp: test.CacheMockArgs{
				Err: cache.ErrNotFound,
			},
			clientErr: errSomethingWrong,
			wantErr:   true,
		},
		{
			name: "err_nil_oldest_storage_time",
			getOp: test.CacheMockArgs{
				Err: cache.ErrNotFound,
			},
			clientResp: &seqapi.StatusResponse{
				OldestStorageTime: nil,
			},
			wantErr: true,
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

			api := setupTestAPI(seqData)
			api.nowFn = func() time.Time {
				return testSomeMoment.Add(result)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getLogsLifespanResponse]{
				Method:  http.MethodGet,
				Target:  "/seqapi/v1/logs_lifespan",
				Handler: api.serveGetLogsLifespan,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
