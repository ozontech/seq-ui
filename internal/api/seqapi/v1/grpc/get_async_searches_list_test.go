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
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeGetAsyncSearchesList(t *testing.T) {
	var (
		mockSearchID1        = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockSearchID2        = "9e4c068e-d4f4-4a5d-be27-a6524a70d70d"
		mockUserName1        = "some_user_1"
		mockUserName2        = "some_user_2"
		mockProfileID1 int64 = 1
		mockProfileID2 int64 = 1
		errorMsg             = "some error"

		mockTime = time.Date(2025, 8, 6, 17, 52, 12, 123, time.UTC)
	)

	type mockArgs struct {
		searchIDs []string

		repoReq  types.GetAsyncSearchesListRequest
		repoResp []types.AsyncSearchInfo
		repoErr  error
	}

	tests := []struct {
		name string

		req  *seqapi.GetAsyncSearchesListRequest
		resp *seqapi.GetAsyncSearchesListResponse
		err  error

		mockArgs *mockArgs
	}{
		{
			name: "ok_no_filters",
			req:  &seqapi.GetAsyncSearchesListRequest{},
			resp: &seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID1,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
							To:        timestamppb.New(mockTime),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
						Error:     &errorMsg,
					},
					{
						SearchId: mockSearchID2,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(360 * time.Second),
							Query:     "message:error and level:3",
							From:      timestamppb.New(mockTime.Add(-1 * time.Hour)),
							To:        timestamppb.New(mockTime),
							Aggs: []*seqapi.AggregationQuery{
								{
									Field:   "x",
									GroupBy: "level",
									Func:    seqapi.AggFunc_AGG_FUNC_AVG,
								},
							},
							Hist: &seqapi.StartAsyncSearchRequest_HistQuery{
								Interval: "1s",
							},
							WithDocs: false,
						},
						StartedAt:  timestamppb.New(mockTime.Add(-60 * time.Second)),
						ExpiresAt:  timestamppb.New(mockTime.Add(300 * time.Second)),
						CanceledAt: timestamppb.New(mockTime),
						Progress:   1,
						DiskUsage:  256,
						OwnerName:  mockUserName2,
					},
				},
			},
			mockArgs: &mockArgs{
				repoReq: types.GetAsyncSearchesListRequest{},
				repoResp: []types.AsyncSearchInfo{
					{
						SearchID:  mockSearchID1,
						OwnerID:   mockProfileID1,
						OwnerName: mockUserName1,
					},
					{
						SearchID:  mockSearchID2,
						OwnerID:   mockProfileID2,
						OwnerName: mockUserName2,
					},
				},
				searchIDs: []string{mockSearchID1, mockSearchID2},
			},
		},
		{
			name: "ok_filters",
			req: &seqapi.GetAsyncSearchesListRequest{
				Status:    seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE.Enum(),
				OwnerName: &mockUserName1,
				Limit:     10,
				Offset:    20,
			},
			resp: &seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: mockSearchID1,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(mockTime.Add(-15 * time.Minute)),
							To:        timestamppb.New(mockTime),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(mockTime.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(mockTime.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
			},
			mockArgs: &mockArgs{
				repoReq: types.GetAsyncSearchesListRequest{
					Owner: &mockUserName1,
				},
				repoResp: []types.AsyncSearchInfo{
					{
						SearchID:  mockSearchID1,
						OwnerID:   mockProfileID1,
						OwnerName: mockUserName1,
					},
				},
				searchIDs: []string{mockSearchID1},
			},
		},
		{
			name: "err_limit",
			req: &seqapi.GetAsyncSearchesListRequest{
				Limit:  -10,
				Offset: 10,
			},
			err: status.Error(codes.InvalidArgument, "invalid request field: 'limit' must be non-negative"),
		},
		{
			name: "err_offset",
			req: &seqapi.GetAsyncSearchesListRequest{
				Limit:  10,
				Offset: -10,
			},
			err: status.Error(codes.InvalidArgument, "invalid request field: 'offset' must be non-negative"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
				asyncSearchesRepoMock.EXPECT().GetAsyncSearchesList(gomock.Any(), tt.mockArgs.repoReq).
					Return(tt.mockArgs.repoResp, tt.mockArgs.repoErr).Times(1)
				seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().GetAsyncSearchesList(gomock.Any(), tt.req, tt.mockArgs.searchIDs).
					Return(tt.resp, nil).Times(1)
				seqData.Mocks.SeqDB = seqDbMock
			}

			api := initTestAPIWithAsyncSearches(seqData)

			ctx := context.Background()

			resp, err := api.GetAsyncSearchesList(ctx, tt.req)
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

func TestServeGetAsyncSearchesList_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)

	_, err := api.GetAsyncSearchesList(context.Background(), &seqapi.GetAsyncSearchesListRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
