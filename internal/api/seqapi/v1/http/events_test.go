package http

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetEvent(t *testing.T) {
	var (
		id1      = "test1"
		id2      = "test2"
		id3      = "test3"
		id4      = "test4"
		cacheTTL = time.Minute
	)

	event1 := test.MakeEvent(id1, 1, testSomeMoment)
	event1json, _ := proto.Marshal(event1)
	event2 := test.MakeEvent(id2, 2, testSomeMoment)
	event2json, _ := proto.Marshal(event2)
	event3 := test.MakeEvent(id3, 0, testSomeMoment)
	event3json, _ := proto.Marshal(event3)

	type mockArgs struct {
		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse
		err  error
	}

	tests := []struct {
		name string

		id      string
		want    getEventResponse
		wantErr bool

		cacheArgs test.CacheMockArgs
		mockArgs  *mockArgs
	}{
		{
			name: "ok_no_cached",
			id:   id1,
			want: getEventResponse{Event: eventFromProto(event1)},
			cacheArgs: test.CacheMockArgs{
				Key:   id1,
				Value: string(event1json),
				Err:   errSomethingWrong,
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetEventRequest{
					Id: id1,
				},
				resp: &seqapi.GetEventResponse{
					Event: event1,
				},
			},
		},
		{
			name: "ok_cached",
			id:   id2,
			want: getEventResponse{Event: eventFromProto(event2)},
			cacheArgs: test.CacheMockArgs{
				Key:   id2,
				Value: string(event2json),
			},
		},
		{
			name: "ok_empty",
			id:   id3,
			want: getEventResponse{Event: eventFromProto(event3)},
			cacheArgs: test.CacheMockArgs{
				Key:   id3,
				Value: string(event3json),
				Err:   errSomethingWrong,
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetEventRequest{
					Id: id3,
				},
				resp: &seqapi.GetEventResponse{
					Event: event3,
				},
			},
		},
		{
			name:    "err_client",
			id:      id4,
			wantErr: true,
			cacheArgs: test.CacheMockArgs{
				Key: id4,
				Err: errSomethingWrong,
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetEventRequest{
					Id: id4,
				},
				err: errors.New("client error"),
			},
		},
	}

	for _, tt := range tests {
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
			cacheMock.EXPECT().
				Get(gomock.Any(), tt.cacheArgs.Key).
				Return(tt.cacheArgs.Value, tt.cacheArgs.Err).
				Times(1)
			seqData.Mocks.Cache = cacheMock

			if tt.mockArgs != nil {
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().
					GetEvent(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				if tt.mockArgs.err == nil {
					cacheMock.EXPECT().
						SetWithTTL(gomock.Any(), tt.cacheArgs.Key, tt.cacheArgs.Value, cacheTTL).
						Return(nil).
						Times(1)
				}
			}

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getEventResponse]{
				Method:  http.MethodGet,
				Target:  fmt.Sprintf("/seqapi/v1/events/%s", tt.id),
				Handler: withQueryParamID(api.serveGetEvent, tt.id),
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestGetEventWithMasking(t *testing.T) {
	var (
		errCache = errors.New("test error")
		cacheTTL = time.Minute
	)
	type mockArgs struct {
		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse
	}

	tests := []struct {
		name string

		shouldMask bool
		isCached   bool
		wantErr    bool

		maskingCfg *config.Masking
	}{
		{
			name:       "mask_noncached",
			shouldMask: true,
			isCached:   false,
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
		want      getEventResponse
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
			Time: timestamppb.New(testSomeMoment),
		}
		if shouldMask {
			event.Data[eventField] = "***"
		}
		eventJson, err := proto.Marshal(event)
		require.NoError(t, err)
		return eventData{
			id:        id,
			event:     event,
			eventJson: eventJson,
			want:      getEventResponse{Event: eventFromProto(event)},
		}
	}

	eventsData := make([]eventData, 0, len(tests))
	for i := 0; i < len(tests); i++ {
		eventsData = append(eventsData, formEventData(i, tests[i].shouldMask))
	}

	for i, tt := range tests {
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
				cacheArgs.Err = errCache
			}
			cacheMock.EXPECT().
				Get(gomock.Any(), cacheArgs.Key).
				Return(cacheArgs.Value, cacheArgs.Err).
				Times(1)
			seqData.Mocks.Cache = cacheMock

			seqDbMock := mock_seqdb.NewMockClient(ctrl)
			if !tt.isCached {
				mockArgs := &mockArgs{
					req:  &seqapi.GetEventRequest{Id: curEData.id},
					resp: &seqapi.GetEventResponse{Event: curEData.event},
				}
				seqDbMock.EXPECT().
					GetEvent(gomock.Any(), mockArgs.req).
					Return(mockArgs.resp, nil).
					Times(1)

				cacheMock.EXPECT().
					SetWithTTL(gomock.Any(), cacheArgs.Key, cacheArgs.Value, cacheTTL).
					Return(nil).
					Times(1)
			}
			if tt.maskingCfg != nil {
				seqDbMock.EXPECT().
					WithMasking(gomock.Any()).
					Return().
					Times(1)
			}
			seqData.Mocks.SeqDB = seqDbMock

			api := setupTestAPI(seqData)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getEventResponse]{
				Method:  http.MethodGet,
				Target:  fmt.Sprintf("/seqapi/v1/events/%s", curEID),
				Handler: withQueryParamID(api.serveGetEvent, curEID),
				Want:    curEData.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
