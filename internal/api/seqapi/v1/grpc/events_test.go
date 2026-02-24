package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						EventsCacheTTL: cacheTTL,
					},
				},
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

			md := metadata.New(map[string]string{"env": "test"})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := s.GetEvent(ctx, tt.req)

			require.Equal(t, tt.clientErr, err)
			if tt.clientErr != nil {
				return
			}

			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}

func TestGetEventWithMasking(t *testing.T) {
	type seqDBArgs struct {
		req  *seqapi.GetEventRequest
		resp *seqapi.GetEventResponse
	}

	eventTime := time.Date(2024, time.December, 31, 10, 20, 30, 400000, time.UTC)

	cacheErr := errors.New("test error")
	cacheTTL := time.Minute

	tests := []struct {
		name string

		shouldMask bool
		isCached   bool
		wantErr    error

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
		wantResp  *seqapi.GetEventResponse
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
		wantResp := &seqapi.GetEventResponse{Event: event}
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

			curEData := eventsData[i]
			curEID := curEData.id

			seqData := test.APITestData{
				Cfg: config.SeqAPI{
					SeqAPIOptions: &config.SeqAPIOptions{
						EventsCacheTTL: cacheTTL,
					},
				},
			}
			ctrl := gomock.NewController(t)

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

			if !tt.isCached {
				seqDBArgs := &seqDBArgs{
					req:  &seqapi.GetEventRequest{Id: curEData.id},
					resp: &seqapi.GetEventResponse{Event: curEData.event},
				}
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetEvent(gomock.Any(), seqDBArgs.req).
					Return(seqDBArgs.resp, nil).Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				cacheMock.EXPECT().SetWithTTL(gomock.Any(), cacheArgs.Key, cacheArgs.Value, cacheTTL).
					Return(nil).Times(1)
			}

			s := initTestAPI(seqData)

			req := &seqapi.GetEventRequest{Id: curEID}

			resp, err := s.GetEvent(context.Background(), req)

			require.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}

			require.True(t, proto.Equal(curEData.wantResp, resp))
		})
	}
}
