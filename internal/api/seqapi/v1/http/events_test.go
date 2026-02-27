package http

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						EventsCacheTTL: cacheTTL,
					},
				},
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

func TestGetEventWithMasking(t *testing.T) {
	type seqDBArgs struct {
		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse
	}

	eventTime := time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC) // 2024-12-31T10:20:30.0004Z

	cacheErr := errors.New("test error")
	cacheTTL := time.Minute

	tests := []struct {
		name string

		shouldMask bool
		isCached   bool
		wantStatus int

		maskingCfg *config.Masking
	}{
		{
			name:       "mask_noncached",
			shouldMask: true,
			isCached:   false,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(val1)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field1",
									Mode:   "equal",
									Values: []string{"val1"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "mask_from_cache",
			shouldMask: true,
			isCached:   true,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(val2)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field2",
									Mode:   "equal",
									Values: []string{"val2"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "do_not_mask_noncached_regex",
			shouldMask: false,
			isCached:   false,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(nomask)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field3",
									Mode:   "equal",
									Values: []string{"val3"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "do_not_mask_from_cache_regex",
			shouldMask: false,
			isCached:   true,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(nomask)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field4",
									Mode:   "equal",
									Values: []string{"val4"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "do_not_mask_noncached_field_filter",
			shouldMask: false,
			isCached:   true,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(val5)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field5",
									Mode:   "equal",
									Values: []string{"val4"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "do_not_mask_from_cache_field_filter",
			shouldMask: false,
			isCached:   true,
			wantStatus: http.StatusOK,
			maskingCfg: &config.Masking{
				Masks: []config.Mask{
					{
						Re:          `(val6)`,
						Groups:      []int{0},
						Mode:        config.MaskModeReplace,
						ReplaceWord: `***`,
						FieldFilters: &config.FieldFilterSet{
							Condition: config.FieldFilterConditionAnd,
							Filters: []config.FieldFilter{
								{
									Field:  "field6",
									Mode:   "equal",
									Values: []string{"val4"},
								},
							},
						},
					},
				},
			},
		},
	}

	type eventData struct {
		id        string
		event     *seqapi.Event
		eventJson []byte
		wantResp  []byte
	}

	formEventData := func(i int, shouldMask bool) eventData {
		num := i + 1
		id := fmt.Sprintf("test%d", num)
		eventField := fmt.Sprintf("field%d", num)
		eventVal := fmt.Sprintf("val%d", num)
		event := &seqapi.Event{
			Id: id,
			Data: map[string]string{
				eventField: eventVal,
			},
			Time: timestamppb.New(eventTime),
		}
		if shouldMask {
			event.Data[eventField] = "***"
		}
		eventJson, err := proto.Marshal(event)
		require.NoError(t, err)
		wantResp, err := json.Marshal(getEventResponse{Event: eventFromProto(event)})
		require.NoError(t, err)
		return eventData{
			id:        id,
			event:     event,
			eventJson: eventJson,
			wantResp:  wantResp,
		}
	}

	eventsData := make([]eventData, 0, len(tests))
	for i := 0; i < len(tests); i++ {
		eventsData = append(eventsData, formEventData(i, tests[i].shouldMask))
	}

	for i, tt := range tests {
		i := i
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			curEData := eventsData[i]
			curEID := curEData.id

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						EventsCacheTTL: cacheTTL,
						Masking:        tt.maskingCfg,
					},
				},
			}

			cacheMock := mock_cache.NewMockCache(ctrl)
			cacheArgs := test.CacheMockArgs{
				Key:   curEID,
				Value: string(curEData.eventJson),
			}
			if !tt.isCached {
				cacheArgs.Err = cacheErr
			}
			cacheMock.EXPECT().Get(gomock.Any(), cacheArgs.Key).
				Return(cacheArgs.Value, cacheArgs.Err).Times(1)
			seqData.Mocks.Cache = cacheMock

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			if !tt.isCached {
				seqDBArgs := &seqDBArgs{
					req:  &seqapi.GetEventRequest{Id: curEData.id},
					resp: &seqapi.GetEventResponse{Event: curEData.event},
				}
				seqDbMock.EXPECT().GetEvent(gomock.Any(), seqDBArgs.req).
					Return(seqDBArgs.resp, nil).Times(1)

				cacheMock.EXPECT().SetWithTTL(gomock.Any(), cacheArgs.Key, cacheArgs.Value, cacheTTL).
					Return(nil).Times(1)
			}
			if tt.maskingCfg != nil {
				seqDbMock.EXPECT().WithMasking(gomock.Any()).Return().Times(1)
			}
			seqData.Mocks.SeqDB = seqDbMock

			api := initTestAPI(seqData)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/seqapi/v1/events/%s", curEID), http.NoBody)
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("id", curEID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetEvent,
				WantRespBody: string(curEData.wantResp),
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
