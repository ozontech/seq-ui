package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeCancelAsyncSearch(t *testing.T) {
	const (
		mockSearchID1  = "69e4a4a6-0922-43bd-952d-060a86c2b622"
		mockUserName1  = "some_user_1"
		mockUserName2  = "some_user_2"
		mockProfileID1 = 1
		mockProfileID2 = 2
	)

	type mockArgs struct {
		userName string

		profilesReq  *types.GetOrCreateUserProfileRequest
		profilesResp *types.UserProfile
		profilesErr  error

		repoResp *types.AsyncSearchInfo
		repoErr  error
	}

	tests := []struct {
		name string

		req  *seqapi.CancelAsyncSearchRequest
		resp *seqapi.CancelAsyncSearchResponse
		err  error

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &seqapi.CancelAsyncSearchRequest{
				SearchId: mockSearchID1,
			},
			resp: &seqapi.CancelAsyncSearchResponse{},
			mockArgs: &mockArgs{
				userName: mockUserName1,
				profilesReq: &types.GetOrCreateUserProfileRequest{
					UserName: mockUserName1,
				},
				profilesResp: &types.UserProfile{
					ID:       mockProfileID1,
					UserName: mockUserName1,
				},
				repoResp: &types.AsyncSearchInfo{
					SearchID:  mockSearchID1,
					OwnerID:   mockProfileID1,
					OwnerName: mockUserName1,
				},
			},
		},
		{
			name: "err_permission_denied",
			req: &seqapi.CancelAsyncSearchRequest{
				SearchId: mockSearchID1,
			},
			mockArgs: &mockArgs{
				userName: mockUserName1,
				profilesReq: &types.GetOrCreateUserProfileRequest{
					UserName: mockUserName1,
				},
				profilesResp: &types.UserProfile{
					ID:       mockProfileID1,
					UserName: mockUserName1,
				},
				repoResp: &types.AsyncSearchInfo{
					SearchID:  mockSearchID1,
					OwnerID:   mockProfileID2,
					OwnerName: mockUserName2,
				},
			},
			err: status.Error(codes.PermissionDenied, "permission denied: cancel async search"),
		},
		{
			name: "invalid id",
			req: &seqapi.CancelAsyncSearchRequest{
				SearchId: "some_invalid_id",
			},
			mockArgs: &mockArgs{
				userName: mockUserName1,
			},
			err: status.Error(codes.InvalidArgument, "invalid search_id"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				if tt.err == nil {
					seqDbMock := mock_seqdb.NewMockClient(ctrl)
					seqDbMock.EXPECT().CancelAsyncSearch(gomock.Any(), tt.req).
						Return(tt.resp, nil).Times(1)
					seqData.Mocks.SeqDB = seqDbMock
				}

				if tt.mockArgs.profilesResp != nil {
					profilesRepoMock := mock_repo.NewMockUserProfiles(ctrl)
					profilesRepoMock.EXPECT().GetOrCreate(gomock.Any(), *tt.mockArgs.profilesReq).
						Return(*tt.mockArgs.profilesResp, tt.mockArgs.profilesErr).Times(1)
					seqData.Mocks.ProfilesRepo = profilesRepoMock
				}

				if tt.mockArgs.repoResp != nil {
					asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
					asyncSearchesRepoMock.EXPECT().GetAsyncSearchById(gomock.Any(), tt.req.SearchId).
						Return(*tt.mockArgs.repoResp, tt.mockArgs.repoErr).Times(1)
					seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock
				}
			}

			api := initTestAPIWithAsyncSearches(seqData)

			ctx := context.Background()
			ctx = context.WithValue(ctx, types.UserKey{}, tt.mockArgs.userName)

			resp, err := api.CancelAsyncSearch(ctx, tt.req)
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

func TestServeCancelAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)

	_, err := api.CancelAsyncSearch(context.Background(), &seqapi.CancelAsyncSearchRequest{})
	require.Error(t, err)
	require.Equal(t, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error()), err)
}
