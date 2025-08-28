package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestGetEvent(t *testing.T) {
	eventTime := time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC)
	id1 := "test1"
	id2 := "test2"
	id3 := "test3"
	id4 := "test4"
	event1 := test.MakeEvent(id1, 1, eventTime)
	event1json, _ := proto.Marshal(event1)
	event2 := test.MakeEvent(id2, 2, eventTime)
	event2json, _ := proto.Marshal(event2)
	event3 := &seqapi.Event{}
	event3json, _ := proto.Marshal(event3)
	err := errors.New("test error")
	cacheTTL := time.Minute

	tests := []struct {
		name string

		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse

		cacheArgs test.CacheMockArgs
		clientErr error
	}{
		{
			name: "ok_no_cached",
			req: &seqapi.GetEventRequest{
				Id: id1,
			},
			resp: &seqapi.GetEventResponse{
				Event: event1,
			},
			cacheArgs: test.CacheMockArgs{
				Key:   id1,
				Value: string(event1json),
				Err:   err,
			},
		},
		{
			name: "ok_cached",
			req: &seqapi.GetEventRequest{
				Id: id2,
			},
			resp: &seqapi.GetEventResponse{
				Event: event2,
			},
			cacheArgs: test.CacheMockArgs{
				Key:   id2,
				Value: string(event2json),
			},
		},
		{
			name: "ok_empty",
			req: &seqapi.GetEventRequest{
				Id: id3,
			},
			resp: &seqapi.GetEventResponse{
				Event: event3,
			},
			cacheArgs: test.CacheMockArgs{
				Key:   id3,
				Value: string(event3json),
				Err:   err,
			},
		},
		{
			name: "err_client",
			req: &seqapi.GetEventRequest{
				Id: id4,
			},
			cacheArgs: test.CacheMockArgs{
				Key: id4,
				Err: err,
			},
			clientErr: errors.New("client error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: config.SeqAPI{EventsCacheTTL: cacheTTL},
			}
			ctrl := gomock.NewController(t)

			cacheMock := mock_cache.NewMockCache(ctrl)
			cacheMock.EXPECT().Get(gomock.Any(), tt.cacheArgs.Key).
				Return(tt.cacheArgs.Value, tt.cacheArgs.Err).Times(1)
			seqData.Mocks.Cache = cacheMock

			if tt.cacheArgs.Err != nil {
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetEvent(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.resp), tt.clientErr).Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				if tt.clientErr == nil {
					cacheMock.EXPECT().SetWithTTL(gomock.Any(), tt.cacheArgs.Key, tt.cacheArgs.Value, cacheTTL).
						Return(nil).Times(1)
				}
			}

			s := initTestAPI(seqData)

			resp, err := s.GetEvent(context.Background(), tt.req)

			require.Equal(t, tt.clientErr, err)
			if tt.clientErr != nil {
				return
			}

			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}
