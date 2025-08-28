package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetFavoriteQueries(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	queryName := "my query"
	var relativeFrom uint64 = 300

	type mockArgs struct {
		req  types.GetFavoriteQueriesRequest
		resp types.FavoriteQueries
		err  error
	}

	tests := []struct {
		name string

		want     *userprofile.GetFavoriteQueriesResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			want: &userprofile.GetFavoriteQueriesResponse{
				Queries: []*userprofile.GetFavoriteQueriesResponse_Query{
					{
						Id:           1,
						Query:        "test1",
						Name:         &queryName,
						RelativeFrom: &relativeFrom,
					},
					{
						Id:           2,
						Query:        "test2",
						Name:         &queryName,
						RelativeFrom: &relativeFrom,
					},
					{
						Id:           3,
						Query:        "test3",
						Name:         &queryName,
						RelativeFrom: &relativeFrom,
					},
					{
						Id:    4,
						Query: "test4",
					},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetFavoriteQueriesRequest{
					ProfileID: profileID,
				},
				resp: types.FavoriteQueries{
					{
						ID:           1,
						Query:        "test1",
						Name:         queryName,
						RelativeFrom: relativeFrom,
					},
					{
						ID:           2,
						Query:        "test2",
						Name:         queryName,
						RelativeFrom: relativeFrom,
					},
					{
						ID:           3,
						Query:        "test3",
						Name:         queryName,
						RelativeFrom: relativeFrom,
					},
					{
						ID:    4,
						Query: "test4",
					},
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name:     "err_repo_random",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetFavoriteQueriesRequest{
					ProfileID: profileID,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newFavoriteQueriesTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetAll(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.GetFavoriteQueries(ctx, &userprofile.GetFavoriteQueriesRequest{})

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestCreateFavoriteQuery(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	var queryID int64 = 1
	query := "test"
	queryName := "my query"
	var relativeFrom uint64 = 300

	type mockArgs struct {
		req  types.GetOrCreateFavoriteQueryRequest
		resp int64
		err  error
	}

	tests := []struct {
		name string

		req      *userprofile.CreateFavoriteQueryRequest
		want     *userprofile.CreateFavoriteQueryResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &userprofile.CreateFavoriteQueryRequest{
				Query:        query,
				Name:         &queryName,
				RelativeFrom: &relativeFrom,
			},
			want: &userprofile.CreateFavoriteQueryResponse{
				Id: queryID,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					ProfileID:    profileID,
					Query:        query,
					Name:         queryName,
					RelativeFrom: relativeFrom,
				},
				resp: queryID,
			},
		},
		{
			name: "success_only_query",
			req: &userprofile.CreateFavoriteQueryRequest{
				Query: query,
			},
			want: &userprofile.CreateFavoriteQueryResponse{
				Id: queryID,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					ProfileID: profileID,
					Query:     query,
				},
				resp: queryID,
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name:     "err_svc_empty_query",
			req:      &userprofile.CreateFavoriteQueryRequest{},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_random",
			req: &userprofile.CreateFavoriteQueryRequest{
				Query: query,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					ProfileID: profileID,
					Query:     query,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newFavoriteQueriesTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.CreateFavoriteQuery(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteFavoriteQuery(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	var queryID int64 = 100

	type mockArgs struct {
		req types.DeleteFavoriteQueryRequest
		err error
	}

	tests := []struct {
		name string

		req      *userprofile.DeleteFavoriteQueryRequest
		want     *userprofile.DeleteFavoriteQueryResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success",
			req: &userprofile.DeleteFavoriteQueryRequest{
				Id: queryID,
			},
			want:     &userprofile.DeleteFavoriteQueryResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID:        queryID,
					ProfileID: profileID,
				},
			},
		},
		{
			name:     "err_no_user",
			wantCode: codes.Unauthenticated,
			noUser:   true,
		},
		{
			name: "err_svc_invalid_id",
			req: &userprofile.DeleteFavoriteQueryRequest{
				Id: -100,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_repo_random",
			req: &userprofile.DeleteFavoriteQueryRequest{
				Id: queryID,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID:        queryID,
					ProfileID: profileID,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newFavoriteQueriesTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Delete(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			if !tt.noUser {
				ctx = context.WithValue(ctx, types.UserKey{}, userName)
				api.profiles.SetID(userName, profileID)
			}

			got, err := api.DeleteFavoriteQuery(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
