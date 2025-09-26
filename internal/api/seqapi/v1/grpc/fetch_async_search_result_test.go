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
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeFetchAsyncSearchResult(t *testing.T) {
	var (
		mockSearchID = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockTime     = time.Date(2025, 8, 6, 17, 52, 12, 123, time.UTC)
	)

	tests := []struct {
		name string

		req  *seqapi.FetchAsyncSearchResultRequest
		resp *seqapi.FetchAsyncSearchResultResponse

		err error
	}{
		{
			name: "ok",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: mockSearchID,
				Limit:    2,
				Offset:   10,
				Order:    seqapi.Order_ORDER_DESC,
			},
			resp: &seqapi.FetchAsyncSearchResultResponse{
				Status: seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
				Request: &seqapi.StartAsyncSearchRequest{
					Retention: durationpb.New(60 * time.Second),
					Query:     "message:error",
					From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
					To:        timestamppb.New(mockTime),
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
							Time: timestamppb.New(mockTime.Add(-1 * time.Minute)),
						},
						{
							Id: "017a854298010000-8502fe7f2aa33df3",
							Data: map[string]string{
								"level":   "2",
								"message": "some error 2",
								"x":       "8",
							},
							Time: timestamppb.New(mockTime.Add(-2 * time.Minute)),
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
									Ts:        timestamppb.New(mockTime),
								},
								{
									Key:       "2",
									Value:     pointerTo(8),
									NotExists: 1,
									Quantiles: []float64{7, 4},
									Ts:        timestamppb.New(mockTime.Add(-1 * time.Minute)),
								},
							},
							NotExists: 2,
						},
					},
				},
				StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
				ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
				Progress:  1,
				DiskUsage: 512,
			},
		},
		{
			name: "invalid id",
			req: &seqapi.FetchAsyncSearchResultRequest{
				SearchId: "some_invalid_id",
			},
			err: status.Error(codes.InvalidArgument, "invalid search_id"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			ctrl := gomock.NewController(t)

			if tt.err == nil {
				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().FetchAsyncSearchResult(gomock.Any(), tt.req).
					Return(tt.resp, nil).Times(1)
				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPIWithAsyncSearches(seqData)

			ctx := context.Background()

			resp, err := api.FetchAsyncSearchResult(ctx, tt.req)
			if tt.err == nil {
				require.NoError(t, err)
				require.True(t, proto.Equal(tt.resp, resp))
			} else {
				require.Error(t, err)
				require.Equal(t, tt.err, err)
			}
		})
	}
}

func TestServeFetchAsyncSearchResult_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)

	_, err := api.FetchAsyncSearchResult(context.Background(), &seqapi.FetchAsyncSearchResultRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}

func pointerTo[T float64](in T) *T {
	return &in
}
