package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetAggregation(t *testing.T) {
	query := "message:error"
	from := time.Now()
	to := from.Add(time.Second)

	tests := []struct {
		name string

		req  *seqapi.GetAggregationRequest
		resp *seqapi.GetAggregationResponse

		apiErr    bool
		clientErr error

		cfg config.SeqAPI
	}{
		{
			name: "ok_multi_agg",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
				},
			},
			resp: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest: 3,
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
				MaxAggregationsPerRequest: 2,
			},
			apiErr: true,
		},
		{
			name: "err_client",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test2"},
				},
			},
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest: 1,
			},
			clientErr: errors.New("client error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if !tt.apiErr {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAggregation(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.resp), tt.clientErr).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			s := initTestAPI(seqData)

			resp, err := s.GetAggregation(context.Background(), tt.req)
			if tt.apiErr {
				require.True(t, err != nil)
				return
			}

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(resp, tt.resp))
		})
	}
}

func TestGetAggregationWithNormalization(t *testing.T) {
	query := "message:error"
	from := time.Now()
	to := from.Add(time.Second)
	bucketUnit := "3s"

	tests := []struct {
		name string

		req             *seqapi.GetAggregationRequest
		resp            *seqapi.GetAggregationResponse
		normalized_resp *seqapi.GetAggregationResponse

		apiErr    bool
		clientErr error

		cfg config.SeqAPI
	}{
		{
			name: "ok_normalize",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1", Func: seqapi.AggFunc_AGG_FUNC_COUNT, BucketUnit: &bucketUnit},
					{Field: "test2", Func: seqapi.AggFunc_AGG_FUNC_COUNT, BucketUnit: &bucketUnit},
				},
			},
			resp: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					BucketUnit: bucketUnit,
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			normalized_resp: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					BucketUnit: bucketUnit,
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
				MaxAggregationsPerRequest:      3,
				DefaultAggregationTsBucketUnit: time.Second,
			},
		},
		{
			name: "ok_normalize_default_bucket_unit",
			req: &seqapi.GetAggregationRequest{
				Query: query,
				From:  timestamppb.New(from),
				To:    timestamppb.New(to),
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1", Func: seqapi.AggFunc_AGG_FUNC_COUNT},
					{Field: "test2", Func: seqapi.AggFunc_AGG_FUNC_COUNT},
				},
			},
			resp: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			normalized_resp: &seqapi.GetAggregationResponse{
				Aggregations: test.MakeAggregations(2, 3, &test.MakeAggOpts{
					BucketUnit: "4s",
					Values: []float64{
						4,
						8,
						12,
					},
				}),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				MaxAggregationsPerRequest:      3,
				DefaultAggregationTsBucketUnit: 4 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{
				Cfg: tt.cfg,
			}

			if !tt.apiErr {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAggregation(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.resp), tt.clientErr).Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			s := initTestAPI(seqData)

			resp, err := s.GetAggregation(context.Background(), tt.req)
			if tt.apiErr {
				require.True(t, err != nil)
				return
			}

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(resp, tt.normalized_resp))
		})
	}
}
