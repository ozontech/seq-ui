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

func TestServeStartAsyncSearch(t *testing.T) {
	type mockArgs struct {
		resp *seqapi.StartAsyncSearchResponse
		err  error
	}

	tests := []struct {
		name string

		req      *seqapi.StartAsyncSearchRequest
		want     *seqapi.StartAsyncSearchResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &seqapi.StartAsyncSearchRequest{
				Retention: durationpb.New(60 * time.Second),
				Query:     query,
				From:      timestamppb.New(from),
				To:        timestamppb.New(to),
				WithDocs:  true,
				Size:      100,
				Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
					Interval: "1s",
				},
				Aggs: []*seqapi.AggregationQuery{
					{
						Field:     "v",
						GroupBy:   "level",
						Func:      seqapi.AggFunc_AGG_FUNC_AVG,
						Quantiles: []float64{0.95},
					},
				},
				Meta: meta,
			},
			want: &seqapi.StartAsyncSearchResponse{
				SearchId: mockSearchID,
			},
			mockArgs: &mockArgs{
				resp: &seqapi.StartAsyncSearchResponse{
					SearchId: mockSearchID,
				},
			},
		},
		{
			name: "err_svc",
			req: &seqapi.StartAsyncSearchRequest{
				Retention: durationpb.New(60 * time.Second),
				Query:     query,
				From:      timestamppb.New(from),
				To:        timestamppb.New(to),
				WithDocs:  true,
				Size:      100,
				Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
					Interval: "1s",
				},
				Aggs: []*seqapi.AggregationQuery{
					{
						Field:     "v",
						GroupBy:   "level",
						Func:      seqapi.AggFunc_AGG_FUNC_AVG,
						Quantiles: []float64{0.95},
					},
				},
				Meta: meta,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
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
					StartAsyncSearch(gomock.Any(), tt.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)

				seqData.Mocks.AsyncSearchesSvc = svcMock
			}

			api := setupAPIWithAsyncSearches(seqData)
			got, err := api.StartAsyncSearch(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServeStartAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupAPI(seqData)

	_, err := api.StartAsyncSearch(context.Background(), &seqapi.StartAsyncSearchRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
