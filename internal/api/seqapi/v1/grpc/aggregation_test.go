package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestGetAggregation(t *testing.T) {
	var (
		query = "message:error"
	)
	tests := []struct {
		name string

		req  *seqapi.GetAggregationRequest
		want *seqapi.GetAggregationResponse

		apiErr    bool
		clientErr error

		cfg config.SeqAPI
	}{
		{
			name: "ok_multi_agg",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(testFrom),
				To:    timestamppb.New(testTo),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
				},
			},
			want: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "err_aggs_limit_max",
			req: &seqapi.GetAggregationRequest{
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
					{Field: "test3"},
				},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 2,
				},
			},
			apiErr: true,
		},
		{
			name: "err_client",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(testFrom),
				To:    timestamppb.New(testTo),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test2"},
				},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 1,
				},
			},
			clientErr: errors.New("client error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if !tt.apiErr {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().
					GetAggregation(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.want), tt.clientErr).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)

			resp, err := api.GetAggregation(context.Background(), tt.req)
			if tt.apiErr {
				require.True(t, err != nil)
				return
			}

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(resp, tt.want))
		})
	}
}

func TestGetAggregationWithNormalization(t *testing.T) {
	var (
		query            = "message:error"
		interval         = "2s"
		targetBucketRate = "3s"
	)
	tests := []struct {
		name string

		req            *seqapi.GetAggregationRequest
		want           *seqapi.GetAggregationResponse
		wantNormalized *seqapi.GetAggregationResponse

		apiErr    bool
		clientErr error

		cfg config.SeqAPI
	}{
		{
			name: "ok_count",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(testFrom),
				To:    timestamppb.New(testTo),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval},
					{Field: "test2", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval},
				},
			},
			want: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					Values: []float64{
						1,
						2,
						3,
					},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			wantNormalized: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					Values: []float64{
						1,
						2,
						3,
					},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
		{
			name: "ok_normalize_count",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(testFrom),
				To:    timestamppb.New(testTo),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval, TargetBucketRate: &targetBucketRate},
					{Field: "test2", Func: seqapi.AggFunc_AGG_FUNC_COUNT, Interval: &interval, TargetBucketRate: &targetBucketRate},
				},
			},
			want: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					TargetBucketRate: targetBucketRate,
					Values: []float64{
						2,
						4,
						6,
					},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			wantNormalized: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					TargetBucketRate: targetBucketRate,
					Values: []float64{
						3,
						6,
						9,
					},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxAggregationsPerRequest: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if !tt.apiErr {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAggregation(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.want), tt.clientErr).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)

			resp, err := api.GetAggregation(context.Background(), tt.req)
			if tt.apiErr {
				require.True(t, err != nil)
				return
			}

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(resp, tt.wantNormalized))
		})
	}
}
