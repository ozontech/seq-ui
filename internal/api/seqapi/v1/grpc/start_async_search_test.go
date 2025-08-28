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

func TestServeStartAsyncSearch(t *testing.T) {
	const (
		mockSearchID  = "c9a34cf8-4c66-484e-9cc2-42979d848656"
		mockUserName  = "some_user"
		mockProfileID = 1
	)

	query := "message:error"
	from := time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	to := from.Add(time.Second)

	type mockArgs struct {
		profilesReq  types.GetOrCreateUserProfileRequest
		profilesResp types.UserProfile
		profilesErr  error

		repoReq types.SaveAsyncSearchRequest
		repoErr error
	}

	tests := []struct {
		name string

		req  *seqapi.StartAsyncSearchRequest
		resp *seqapi.StartAsyncSearchResponse

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
			},
			resp: &seqapi.StartAsyncSearchResponse{
				SearchId: mockSearchID,
			},
			mockArgs: &mockArgs{
				profilesReq: types.GetOrCreateUserProfileRequest{
					UserName: mockUserName,
				},
				profilesResp: types.UserProfile{
					ID:       mockProfileID,
					UserName: mockUserName,
				},
				repoReq: types.SaveAsyncSearchRequest{
					SearchID: mockSearchID,
					OwnerID:  mockProfileID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				seqDbMock := mock_seqdb.NewMockClient(ctrl)
				seqDbMock.EXPECT().StartAsyncSearch(gomock.Any(), tt.req).
					Return(tt.resp, nil).Times(1)
				seqData.Mocks.SeqDB = seqDbMock

				profilesRepoMock := mock_repo.NewMockUserProfiles(ctrl)
				profilesRepoMock.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.profilesReq).
					Return(tt.mockArgs.profilesResp, tt.mockArgs.profilesErr).Times(1)
				seqData.Mocks.ProfilesRepo = profilesRepoMock

				asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
				asyncSearchesRepoMock.EXPECT().SaveAsyncSearch(gomock.Any(), gomock.Any()).
					Return(tt.mockArgs.repoErr).Times(1)
				seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock
			}

			api := initTestAPIWithAsyncSearches(seqData)

			ctx := context.Background()
			ctx = context.WithValue(ctx, types.UserKey{}, mockUserName)

			resp, err := api.StartAsyncSearch(ctx, tt.req)
			require.NoError(t, err)

			require.True(t, proto.Equal(tt.resp, resp))
		})
	}
}

func TestServeStartAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)

	_, err := api.StartAsyncSearch(context.Background(), &seqapi.StartAsyncSearchRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
