package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeFetchAsyncSearchResult(t *testing.T) {
	type mockArgs struct {
		req *seqapi.FetchAsyncSearchResultRequest
		err error
	}

	tests := []struct {
		name string

		req      *seqapi.FetchAsyncSearchResultRequest
		want     *seqapi.FetchAsyncSearchResultResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: mockSearchID,
				Limit:    2,
				Offset:   10,
				Order:    seqapi.Order_ORDER_DESC,
			},
			want: &seqapi.FetchAsyncSearchResultResponse{
				Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
				Request: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     "message:error",
					From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
					To:        timestamppb.New(someMoment),
					Aggs: []*seqapi.AggregationQuery{
						{
							Field:     "x",
							GroupBy:   "level",
							Func:      seqapi.AggFunc_AGG_FUNC_AVG,
							Quantiles: []float64{0.9, 0.5},
						},
					},
					Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
						Interval: "1s",
					},
					WithDocs: true,
					Size:     100,
				},
				Response: &seqapi.SearchResponse{
					Events: []*seqapi.Event{
						{
							Id: "017a854298010000-850287cfa326a7fc",
							Data: map[string]string{
								"level":   "3",
								"message": "some error",
								"x":       "2",
							},
							Time: timestamppb.New(someMoment.Add(-1 * time.Minute)),
						},
						{
							Id: "017a854298010000-8502fe7f2aa33df3",
							Data: map[string]string{
								"level":   "2",
								"message": "some error 2",
								"x":       "8",
							},
							Time: timestamppb.New(someMoment.Add(-2 * time.Minute)),
						},
					},
					Total: 2,
					Histogram: &seqapi.Histogram{
						Buckets: []*seqapi.Histogram_Bucket{
							{
								DocCount: 7,
								Key:      1,
							},
							{
								DocCount: 9,
								Key:      2,
							},
						},
					},
					Aggregations: []*seqapi.Aggregation{
						{
							Buckets: []*seqapi.Aggregation_Bucket{
								{
									Key:       "3",
									Value:     pointerTo(2),
									NotExists: 0,
									Quantiles: []float64{2, 1},
									Ts:        timestamppb.New(someMoment),
								},
								{
									Key:       "2",
									Value:     pointerTo(8),
									NotExists: 1,
									Quantiles: []float64{7, 4},
									Ts:        timestamppb.New(someMoment.Add(-1 * time.Minute)),
								},
							},
							NotExists: 2,
						},
					},
					Error: &seqapi.Error{
						Code:    seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED,
						Message: "some error",
					},
				},
				StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
				ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
				Progress:  1,
				DiskUsage: 512,
				Meta:      meta,
			},
			mockArgs: &mockArgs{
				req: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: mockSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
			},
		},
		{
			name: "partial_response",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: mockSearchID,
				Limit:    2,
				Offset:   10,
				Order:    seqapi.Order_ORDER_DESC,
			},
			want: &seqapi.FetchAsyncSearchResultResponse{
				Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
				Request: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     "message:error",
					From:      timestamppb.New(someMoment.Add(-15 * time.Minute)),
					To:        timestamppb.New(someMoment),
					WithDocs:  true,
					Size:      100,
				},
				Response: &seqapi.SearchResponse{
					Events: []*seqapi.Event{
						{
							Id: "017a854298010000-850287cfa326a7fc",
							Data: map[string]string{
								"level":   "3",
								"message": "some error",
								"x":       "2",
							},
							Time: timestamppb.New(someMoment.Add(-1 * time.Minute)),
						},
					},
					Total: 1,
					Error: &seqapi.Error{
						Code: seqapi.ErrorCode_ERROR_CODE_NO,
					},
				},
				StartedAt: timestamppb.New(someMoment.Add(-30 * time.Second)),
				ExpiresAt: timestamppb.New(someMoment.Add(30 * time.Second)),
				Progress:  1,
				DiskUsage: 512,
				Meta:      meta,
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
					Message: "partial response",
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: mockSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
			},
		},
		{
			name: "invalid_id",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: "some_invalid_id",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: mockSearchID,
				Limit:    2,
				Offset:   10,
				Order:    seqapi.Order_ORDER_DESC,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: &seqapi.FetchAsyncSearchResultRequest{
					SearchId: mockSearchID,
					Limit:    2,
					Offset:   10,
					Order:    seqapi.Order_ORDER_DESC,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)
				svcMock := mock_asyncsearches.NewMockService(ctrl)

				svcMock.EXPECT().
					FetchAsyncSearchResult(gomock.Any(), tt.mockArgs.req).
					Return(tt.want, tt.mockArgs.err).
					Times(1)
				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupAPIWithAsyncSearches(seqData)
			got, err := api.FetchAsyncSearchResult(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServeFetchAsyncSearchResult_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupAPI(seqData)

	_, err := api.FetchAsyncSearchResult(context.Background(), &seqapi.FetchAsyncSearchResultRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}

func pointerTo[T float64](in T) *T {
	return &in
}
