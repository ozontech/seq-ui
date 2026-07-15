package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
)

func TestGetFavoriteQueries(t *testing.T) {
	var (
		relativeFrom uint64 = 300
		queryName           = "my query"
	)

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
	}{
		{
			name: "ok",
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
			name:     "err_svc",
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetFavoriteQueries(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.GetFavoriteQueries(context.Background(), &userprofile.GetFavoriteQueriesRequest{})

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestCreateFavoriteQuery(t *testing.T) {
	var (
		queryID      int64  = 1
		relativeFrom uint64 = 300
		query               = "test"
		queryName           = "my query"
	)

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
	}{
		{
			name: "ok",
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
					Query:        query,
					Name:         queryName,
					RelativeFrom: relativeFrom,
				},
				resp: queryID,
			},
		},
		{
			name: "err_svc",
			req: &userprofile.CreateFavoriteQueryRequest{
				Query: query,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{Query: query},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetOrCreateFavoriteQuery(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.CreateFavoriteQuery(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteFavoriteQuery(t *testing.T) {
	var (
		queryID int64 = 1
	)

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
	}{
		{
			name: "ok",
			req: &userprofile.DeleteFavoriteQueryRequest{
				Id: queryID,
			},
			want:     &userprofile.DeleteFavoriteQueryResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID: queryID,
				},
			},
		},
		{
			name: "err_svc",
			req: &userprofile.DeleteFavoriteQueryRequest{
				Id: queryID,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID: queryID,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					DeleteFavoriteQuery(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			got, err := api.DeleteFavoriteQuery(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
