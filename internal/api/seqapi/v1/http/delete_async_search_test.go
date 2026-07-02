package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func TestServeDeleteAsyncSearch(t *testing.T) {
	const (
		mockSearchID1  = "69e4a4a6-0922-43bd-952d-060a86c2b622"
		mockUserName1  = "some_user_1"
		mockUserName2  = "some_user_2"
		mockProfileID1 = 1
		mockProfileID2 = 2
	)

	type mockArgs struct {
		userName string

		proxyReq  *seqapi.DeleteAsyncSearchRequest
		proxyResp *seqapi.DeleteAsyncSearchResponse
		proxyErr  error

		profilesReq  *types.GetOrCreateUserProfileRequest
		profilesResp *types.UserProfile
		profilesErr  error

		repoGetAsyncSearchResp *types.AsyncSearchInfo
		repoGetAsyncSearchErr  error

		repoDeleteAsyncSearchErr error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int
		shouldDelete bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			mockArgs: &mockArgs{
				userName: mockUserName1,
				proxyReq: &seqapi.DeleteAsyncSearchRequest{
					SearchId: mockSearchID1,
				},
				proxyResp: &seqapi.DeleteAsyncSearchResponse{},
				profilesReq: &types.GetOrCreateUserProfileRequest{
					UserName: mockUserName1,
				},
				profilesResp: &types.UserProfile{
					ID:       mockProfileID1,
					UserName: mockUserName1,
				},
				repoGetAsyncSearchResp: &types.AsyncSearchInfo{
					SearchID:  mockSearchID1,
					OwnerID:   mockProfileID1,
					OwnerName: mockUserName1,
				},
			},
			shouldDelete: true,
			wantRespBody: ``,
			wantStatus:   http.StatusOK,
		},
		{
			name: "err_permission_denied",
			mockArgs: &mockArgs{
				userName: mockUserName1,
				proxyReq: &seqapi.DeleteAsyncSearchRequest{
					SearchId: mockSearchID1,
				},
				profilesReq: &types.GetOrCreateUserProfileRequest{
					UserName: mockUserName1,
				},
				profilesResp: &types.UserProfile{
					ID:       mockProfileID1,
					UserName: mockUserName1,
				},
				repoGetAsyncSearchResp: &types.AsyncSearchInfo{
					SearchID:  mockSearchID1,
					OwnerID:   mockProfileID2,
					OwnerName: mockUserName2,
				},
			},
			wantRespBody: `{"message":"permission denied: delete async search"}`,
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name: "invalid id",
			mockArgs: &mockArgs{
				userName: mockUserName1,
				proxyReq: &seqapi.DeleteAsyncSearchRequest{
					SearchId: "some_invalid_id",
				},
			},
			wantRespBody: `{"message":"invalid request field: invalid uuid"}`,
			wantStatus:   http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seqData := test.APITestData{}

			if tt.mockArgs != nil {
				ctrl := gomock.NewController(t)

				if tt.mockArgs.proxyResp != nil {
					seqDbMock := mock_seqdb.NewMockClient(ctrl)
					seqDbMock.EXPECT().DeleteAsyncSearch(gomock.Any(), tt.mockArgs.proxyReq).
						Return(tt.mockArgs.proxyResp, tt.mockArgs.proxyErr).Times(1)
					seqData.Mocks.SeqDB = seqDbMock
				}

				if tt.mockArgs.profilesResp != nil {
					profilesRepoMock := mock_repo.NewMockUserProfiles(ctrl)
					profilesRepoMock.EXPECT().GetOrCreate(gomock.Any(), *tt.mockArgs.profilesReq).
						Return(*tt.mockArgs.profilesResp, tt.mockArgs.profilesErr).Times(1)
					seqData.Mocks.ProfilesRepo = profilesRepoMock
				}

				if tt.mockArgs.repoGetAsyncSearchResp != nil {
					asyncSearchesRepoMock := mock_repo.NewMockAsyncSearches(ctrl)
					asyncSearchesRepoMock.EXPECT().GetAsyncSearchById(gomock.Any(), tt.mockArgs.proxyReq.SearchId).
						Return(*tt.mockArgs.repoGetAsyncSearchResp, tt.mockArgs.repoGetAsyncSearchErr).Times(1)

					if tt.shouldDelete {
						asyncSearchesRepoMock.EXPECT().DeleteAsyncSearch(gomock.Any(), tt.mockArgs.proxyReq.SearchId).
							Return(tt.mockArgs.repoDeleteAsyncSearchErr).Times(1)
					}

					seqData.Mocks.AsyncSearchesRepo = asyncSearchesRepoMock
				}
			}

			api := initTestAPIWithAsyncSearches(seqData)
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/seqapi/v1/async_search/%s", tt.mockArgs.proxyReq.SearchId),
				http.NoBody,
			)
			req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, tt.mockArgs.userName))
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("id", tt.mockArgs.proxyReq.SearchId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveDeleteAsyncSearch,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeDeleteAsyncSearch_Disabled(t *testing.T) {
	seqData := test.APITestData{}
	api := initTestAPI(seqData)
	req := httptest.NewRequest(
		http.MethodDelete,
		"/seqapi/v1/async_search/c9a34cf8-4c66-484e-9cc2-42979d848656",
		http.NoBody,
	)

	httputil.DoTestHTTP(t, httputil.TestDataHTTP{
		Req:          req,
		Handler:      api.serveDeleteAsyncSearch,
		WantRespBody: `{"message":"async searches disabled"}`,
		WantStatus:   http.StatusBadRequest,
	})
}
