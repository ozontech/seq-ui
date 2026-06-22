package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestSearch(t *testing.T) {
	var (
		query       = "message:error"
		limit int32 = 3
	)
	tests := []struct {
		name string

		req      *seqapi.SearchRequest
		resp     *seqapi.SearchResponse
		wantResp *seqapi.SearchResponse

		apiErr    bool
		clientErr error

		cfg config.SeqAPI
	}{
		{
			name: "ok",
			req: &seqapi.SearchRequest{
				Query:     query,
				From:      timestamppb.New(testFrom),
				To:        timestamppb.New(testTo),
				Limit:     limit,
				Offset:    0,
				WithTotal: true,
				Histogram: &seqapi.SearchRequest_Histogram{
					Interval: "5s",
				},
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
				},
				Order: seqapi.Order_ORDER_ASC,
			},
			resp: &seqapi.SearchResponse{
				Events:       test.MakeEvents(int(limit), testSomeMoment),
				Total:        int64(limit),
				Histogram:    test.MakeHistogram(2),
				Aggregations: test.MakeAggregations(2, 2, nil),
				Error: &seqapi.Error{
					Code: seqapi.ErrorCode_ERROR_CODE_NO,
				},
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:            5,
					MaxAggregationsPerRequest: 5,
				},
			}),
		},
		{
			name: "err_search_limit_zero",
			req: &seqapi.SearchRequest{
				Limit: 0,
			},
			apiErr: true,
		},
		{
			name: "err_search_limit_max",
			req: &seqapi.SearchRequest{
				Limit: 10,
			},
			cfg: config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			},
			apiErr: true,
		},
		{
			name: "err_aggs_limit_max",
			req: &seqapi.SearchRequest{
				Limit: 3,
				Aggregations: []*seqapi.AggregationQuery{
					{Field: "test1"},
					{Field: "test2"},
					{Field: "test3"},
				},
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:            5,
					MaxAggregationsPerRequest: 2,
				},
			}),
			apiErr: true,
		},
		{
			name: "err_offset_too_high",
			req: &seqapi.SearchRequest{
				Query:  query,
				From:   timestamppb.New(testFrom),
				To:     timestamppb.New(testTo),
				Limit:  limit,
				Offset: 11,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:       5,
					MaxSearchOffsetLimit: 10,
				},
			}),
			apiErr: true,
		},
		{
			name: "err_total_too_high",
			req: &seqapi.SearchRequest{
				Query:     query,
				From:      timestamppb.New(testFrom),
				To:        timestamppb.New(testTo),
				Limit:     limit,
				Offset:    0,
				WithTotal: true,
			},
			resp: &seqapi.SearchResponse{
				Events: test.MakeEvents(int(limit), testSomeMoment),
				Total:  int64(limit) + 1,
			},
			wantResp: &seqapi.SearchResponse{
				Events: test.MakeEvents(int(limit), testSomeMoment),
				Total:  int64(limit) + 1,
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY,
					Message: api_error.ErrQueryTooHeavy.Error(),
				},
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit:      5,
					MaxSearchTotalLimit: int64(limit),
				},
			}),
		},
		{
			name: "err_client",
			req: &seqapi.SearchRequest{
				Query:  query,
				From:   timestamppb.New(testFrom),
				To:     timestamppb.New(testTo),
				Limit:  limit,
				Offset: 0,
			},
			cfg: test.SetCfgDefaults(config.SeqAPI{
				SeqAPIOptions: &config.SeqAPIOptions{
					MaxSearchLimit: 5,
				},
			}),
			clientErr: errSomethingWrong,
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
					Search(gomock.Any(), proto.Clone(tt.req)).
					Return(proto.Clone(tt.resp), tt.clientErr).
					Times(1)

				seqData.Mocks.SeqDB = seqDbMock
			}

			api := setupTestAPI(seqData)
			resp, err := api.Search(context.Background(), tt.req)
			if tt.apiErr {
				require.NotNil(t, err)
				return
			}

			require.Equal(t, tt.clientErr, err)
			if tt.wantResp != nil {
				require.True(t, proto.Equal(tt.wantResp, resp))
			} else {
				require.True(t, proto.Equal(tt.resp, resp))
			}
		})
	}
}
