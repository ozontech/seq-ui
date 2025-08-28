package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestServeGetEvent(t *testing.T) {
	type seqDBArgs struct {
		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse
		err  error
	}

	eventTime := time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC) // 2024-12-31T10:20:30.0004Z
	id1 := "test1"
	id2 := "test2"
	id3 := "test3"
	id4 := "test4"
	event1 := test.MakeEvent(id1, 1, eventTime)
	event1json, _ := proto.Marshal(event1)
	event2 := test.MakeEvent(id2, 2, eventTime)
	event2json, _ := proto.Marshal(event2)
	event3 := test.MakeEvent(id3, 0, eventTime)
	event3json, _ := proto.Marshal(event3)
	err := errors.New("test error")
	cacheTTL := time.Minute

	tests := []struct {
		name string

		id           string
		wantRespBody string
		wantStatus   int

		cacheArgs test.CacheMockArgs
		seqDBArgs *seqDBArgs
	}{
		{
			name: "ok_no_cached",
			id:   id1,
			cacheArgs: test.CacheMockArgs{
				Key:   id1,
				Value: string(event1json),
				Err:   err,
			},
			seqDBArgs: &seqDBArgs{
				req: &seqapi.GetEventRequest{
					Id: id1,
				},
				resp: &seqapi.GetEventResponse{
					Event: event1,
				},
			},
			wantRespBody: `{"event":{"id":"test1","data":{"field1":"val1"},"time":"2024-12-31T10:20:30.0004Z"}}`,
			wantStatus:   http.StatusOK,
		},
		{
			name: "ok_cached",
			id:   id2,
			cacheArgs: test.CacheMockArgs{
				Key:   id2,
				Value: string(event2json),
			},
			wantRespBody: `{"event":{"id":"test2","data":{"field1":"val1","field2":"val2"},"time":"2024-12-31T10:20:30.0004Z"}}`,
			wantStatus:   http.StatusOK,
		},
		{
			name: "ok_empty",
			id:   id3,
			cacheArgs: test.CacheMockArgs{
				Key:   id3,
				Value: string(event3json),
				Err:   err,
			},
			seqDBArgs: &seqDBArgs{
				req: &seqapi.GetEventRequest{
					Id: id3,
				},
				resp: &seqapi.GetEventResponse{
					Event: event3,
				},
			},
			wantRespBody: `{"event":{"id":"test3","data":{},"time":"2024-12-31T10:20:30.0004Z"}}`,
			wantStatus:   http.StatusOK,
		},
		{
			name: "err_client",
			id:   id4,
			cacheArgs: test.CacheMockArgs{
				Key: id4,
				Err: err,
			},
			seqDBArgs: &seqDBArgs{
				req: &seqapi.GetEventRequest{
					Id: id4,
				},
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

			seqData := test.APITestData{
				Cfg: config.SeqAPI{EventsCacheTTL: cacheTTL},
			}

			cacheMock := mock_cache.NewMockCache(ctrl)
			cacheMock.EXPECT().Get(gomock.Any(), tt.cacheArgs.Key).
				Return(tt.cacheArgs.Value, tt.cacheArgs.Err).Times(1)
			seqData.Mocks.Cache = cacheMock

			if tt.seqDBArgs != nil {
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetEvent(gomock.Any(), tt.seqDBArgs.req).
					Return(tt.seqDBArgs.resp, tt.seqDBArgs.err).Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				if tt.seqDBArgs.err == nil {
					cacheMock.EXPECT().SetWithTTL(gomock.Any(), tt.cacheArgs.Key, tt.cacheArgs.Value, cacheTTL).
						Return(nil).Times(1)
				}
			}

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/seqapi/v1/events/%s", tt.id), http.NoBody)
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetEvent,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
