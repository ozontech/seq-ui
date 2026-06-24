package http

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetFavoriteQueries(t *testing.T) {
	var (
		relativeFrom = "300"
	)

	type mockArgs struct {
		req  types.GetFavoriteQueriesRequest
		resp types.FavoriteQueries
		err  error
	}

	tests := []struct {
		name string

		want    getFavoriteQueriesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			want: getFavoriteQueriesResponse{
				Queries: favoriteQueries{
					{ID: "1", Query: "test1", Name: "my query 1", RelativeFrom: relativeFrom},
					{ID: "2", Query: "test2", Name: "my query 2"},
					{ID: "3", Query: "test3", RelativeFrom: relativeFrom},
					{ID: "4", Query: "test4"},
				},
			},
			mockArgs: &mockArgs{
				resp: types.FavoriteQueries{
					{ID: 1, Query: "test1", Name: "my query 1", RelativeFrom: 300},
					{ID: 2, Query: "test2", Name: "my query 2"},
					{ID: 3, Query: "test3", RelativeFrom: 300},
					{ID: 4, Query: "test4"},
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getFavoriteQueriesResponse]{
				Method:  http.MethodGet,
				Target:  "/userprofile/v1/queries/favorite",
				Handler: api.serveGetFavoriteQueries,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeCreateFavoriteQuery(t *testing.T) {
	var (
		relativeFrom = "300"
		query        = "test"
		queryName    = "my query"
	)

	type mockArgs struct {
		req  types.GetOrCreateFavoriteQueryRequest
		resp int64
		err  error
	}

	tests := []struct {
		name string

		req     createFavoriteQueryRequest
		want    createFavoriteQueryResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  createFavoriteQueryRequest{Query: query, Name: &queryName, RelativeFrom: &relativeFrom},
			want: createFavoriteQueryResponse{ID: "1"},
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					Query:        query,
					Name:         "my query",
					RelativeFrom: 300,
				},
				resp: 1,
			},
		},
		{
			name:    "err_svc",
			req:     createFavoriteQueryRequest{Query: query, Name: &queryName, RelativeFrom: &relativeFrom},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					Query:        query,
					Name:         "my query",
					RelativeFrom: 300,
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
					GetOrCreateFavoriteQuery(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[createFavoriteQueryRequest, createFavoriteQueryResponse]{
				Method:  http.MethodPost,
				Target:  "/userprofile/v1/queries/favorite",
				Req:     tt.req,
				Handler: api.serveCreateFavoriteQuery,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}

func TestServeDeleteFavoriteQuery(t *testing.T) {
	type mockArgs struct {
		req types.DeleteFavoriteQueryRequest
		err error
	}

	tests := []struct {
		name string

		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID: 100,
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID: 100,
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

			id := strconv.FormatInt(tt.mockArgs.req.ID, 10)
			handler := withID(api.serveDeleteFavoriteQuery, id)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, struct{}]{
				Method:  http.MethodDelete,
				Target:  fmt.Sprintf("/userprofile/v1/queries/favorite/%s", id),
				Handler: handler,
				NoResp:  true,
				WantErr: tt.wantErr,
			})
		})
	}
}
