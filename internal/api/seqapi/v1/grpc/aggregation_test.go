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
	"google.golang.org/grpc/metadata"
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
			name: "ok_single_agg",
			req: &seqapi.GetAggregationRequest{
				Query:    query,
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				AggField: "test1",
			},
			resp: &seqapi.GetAggregationResponse{
				Aggregation:  test.MakeAggregation(2, nil),
				Aggregations: test.MakeAggregations(1, 2, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB: "test",
						Options: &config.SeqAPIOptions{
							MaxAggregationsPerRequest: 4,
						},
					},
				},
				DefaultEnv: "test",
			},
		},
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
				Aggregation:  test.MakeAggregation(3, nil),
				Aggregations: test.MakeAggregations(2, 3, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB: "test",
						Options: &config.SeqAPIOptions{
							MaxAggregationsPerRequest: 3,
						},
					},
				},
				DefaultEnv: "test",
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
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB: "test",
						Options: &config.SeqAPIOptions{
							MaxAggregationsPerRequest: 2,
						},
					},
				},
				DefaultEnv: "test",
			},
			apiErr: true,
		},
		{
			name: "err_client",
			req: &seqapi.GetAggregationRequest{
				Query:    query,
				From:     timestamppb.New(from),
				To:       timestamppb.New(to),
				AggField: "test2",
			},
			cfg: config.SeqAPI{
				Envs: map[string]config.SeqAPIEnv{
					"test": {
						SeqDB: "test",
						Options: &config.SeqAPIOptions{
							MaxAggregationsPerRequest: 1,
						},
					},
				},
				DefaultEnv: "test",
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

			md := metadata.New(map[string]string{"env": "test"})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := s.GetAggregation(ctx, tt.req)
			if tt.apiErr {
				require.True(t, err != nil)
				return
			}

			require.Equal(t, tt.clientErr, err)
			require.True(t, proto.Equal(resp, tt.resp))
		})
	}
}
