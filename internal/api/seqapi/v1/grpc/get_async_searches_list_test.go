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

func TestServeGetAsyncSearchesList(t *testing.T) {
	var (
		errorMsg      = "some err"
		mockUserName1 = "some_user_1"
		mockUserName2 = "some_user_2"
		mockSearchID2 = "9e4c068e-d4f4-4a5d-be27-a6524a70d70d"
	)
	type mockArgs struct {
		req *seqapi.GetAsyncSearchesListRequest
		err error
	}

	tests := []struct {
		name string

		req      *seqapi.GetAsyncSearchesListRequest
		want     *seqapi.GetAsyncSearchesListResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok_no_filters",
			req:  &seqapi.GetAsyncSearchesListRequest{},
			want: &seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: testSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(testTimestamp.Add(-15 * time.Minute)),
							To:        timestamppb.New(testTimestamp),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(testTimestamp.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(testTimestamp.Add(30 * time.Second)),
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
							From:      timestamppb.New(testTimestamp.Add(-1 * time.Hour)),
							To:        timestamppb.New(testTimestamp),
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
						StartedAt:  timestamppb.New(testTimestamp.Add(-60 * time.Second)),
						ExpiresAt:  timestamppb.New(testTimestamp.Add(300 * time.Second)),
						CanceledAt: timestamppb.New(testTimestamp),
						Progress:   1,
						DiskUsage:  256,
						OwnerName:  mockUserName2,
					},
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{},
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
			want: &seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: testSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(testTimestamp.Add(-15 * time.Minute)),
							To:        timestamppb.New(testTimestamp),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(testTimestamp.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(testTimestamp.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{
					Status:    seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE.Enum(),
					OwnerName: &mockUserName1,
					Limit:     10,
					Offset:    20,
				},
			},
		},
		{
			name: "partial_response",
			req:  &seqapi.GetAsyncSearchesListRequest{},
			want: &seqapi.GetAsyncSearchesListResponse{
				Searches: []*seqapi.GetAsyncSearchesListResponse_ListItem{
					{
						SearchId: testSearchID,
						Status:   seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE,
						Request: &seqapi.StartAsyncSearchRequest{
							Retention: durationpb.New(60 * time.Second),
							Query:     "message:error",
							From:      timestamppb.New(testTimestamp.Add(-15 * time.Minute)),
							To:        timestamppb.New(testTimestamp),
							WithDocs:  true,
							Size:      100,
						},
						StartedAt: timestamppb.New(testTimestamp.Add(-30 * time.Second)),
						ExpiresAt: timestamppb.New(testTimestamp.Add(30 * time.Second)),
						Progress:  1,
						DiskUsage: 512,
						OwnerName: mockUserName1,
					},
				},
				Error: &seqapi.Error{
					Code:    seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE,
					Message: "partial response",
				},
			},
			mockArgs: &mockArgs{
				req: &seqapi.GetAsyncSearchesListRequest{},
			},
		},
		{
			name: "err_limit",
			req: &seqapi.GetAsyncSearchesListRequest{
				Limit:  -10,
				Offset: 10,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_offset",
			req: &seqapi.GetAsyncSearchesListRequest{
				Limit:  10,
				Offset: -10,
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			svcMock := mock_asyncsearches.NewMockService(ctrl)

			seqData := test.APITestData{}
			seqData.Mocks.AsyncSearchesSvc = svcMock

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					GetAsyncSearchesList(gomock.Any(), tt.mockArgs.req).
					Return(tt.want, tt.mockArgs.err).
					Times(1)
			}

			api := setupTestAPI(seqData)
			got, err := api.GetAsyncSearchesList(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServeGetAsyncSearchesList_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := setupTestAPI(seqData)

	_, err := api.GetAsyncSearchesList(context.Background(), &seqapi.GetAsyncSearchesListRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
