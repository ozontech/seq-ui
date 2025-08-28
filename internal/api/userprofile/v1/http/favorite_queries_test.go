package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"go.uber.org/mock/gomock"
)

func TestServeGetFavoriteQueries(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1

	type mockArgs struct {
		req  types.GetFavoriteQueriesRequest
		resp types.FavoriteQueries
		err  error
	}

	tests := []struct {
		name string

		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:         "success",
			wantRespBody: `{"queries":[{"id":"1","query":"test1","name":"my query 1","relativeFrom":"300"},{"id":"2","query":"test2","name":"my query 2"},{"id":"3","query":"test3","relativeFrom":"900"},{"id":"4","query":"test4"}]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetFavoriteQueriesRequest{
					ProfileID: profileID,
				},
				resp: types.FavoriteQueries{
					{
						ID:           1,
						Query:        "test1",
						Name:         "my query 1",
						RelativeFrom: 300,
					},
					{
						ID:    2,
						Query: "test2",
						Name:  "my query 2",
					},
					{
						ID:           3,
						Query:        "test3",
						RelativeFrom: 900,
					},
					{
						ID:    4,
						Query: "test4",
					},
				},
			},
		},
		{
			name:       "err_no_user",
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_repo_random",
			wantStatus: http.StatusInternalServerError,
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
			req := httptest.NewRequest(http.MethodGet, "/userprofile/v1/queries/favorite", http.NoBody)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetAll(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetFavoriteQueries,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeCreateFavoriteQuery(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	var queryID int64 = 1
	query := "test"

	formatReqBody := func(query, name, relativeFrom string) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`{"query":%q`, query))
		if name != "" {
			sb.WriteString(fmt.Sprintf(`,"name":%q`, name))
		}
		if relativeFrom != "" {
			sb.WriteString(fmt.Sprintf(`,"relativeFrom":%q`, relativeFrom))
		}
		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req  types.GetOrCreateFavoriteQueryRequest
		resp int64
		err  error
	}

	tests := []struct {
		name string

		reqBody      string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:         "success",
			reqBody:      formatReqBody(query, "my query", "300"),
			wantRespBody: fmt.Sprintf(`{"id":"%d"}`, queryID),
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					ProfileID:    profileID,
					Query:        query,
					Name:         "my query",
					RelativeFrom: 300,
				},
				resp: queryID,
			},
		},
		{
			name:         "success_only_query",
			reqBody:      formatReqBody(query, "", ""),
			wantRespBody: fmt.Sprintf(`{"id":"%d"}`, queryID),
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetOrCreateFavoriteQueryRequest{
					ProfileID: profileID,
					Query:     query,
				},
				resp: queryID,
			},
		},
		{
			name:       "err_invalid_request",
			reqBody:    "invalid-request",
			wantStatus: http.StatusBadRequest,
			noUser:     true,
		},
		{
			name:       "err_no_user",
			reqBody:    formatReqBody(query, "", ""),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_invalid_relative_from_format",
			reqBody:    formatReqBody(query, "", "not_number"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_empty_query",
			reqBody:    formatReqBody("", "", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_random",
			reqBody:    formatReqBody(query, "", ""),
			wantStatus: http.StatusInternalServerError,
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
			req := httptest.NewRequest(http.MethodPost, "/userprofile/v1/queries/favorite", strings.NewReader(tt.reqBody))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetOrCreate(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveCreateFavoriteQuery,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}

func TestServeDeleteFavoriteQuery(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1

	type mockArgs struct {
		req types.DeleteFavoriteQueryRequest
		err error
	}

	tests := []struct {
		name string

		id         string
		wantStatus int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:       "success",
			id:         "100",
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID:        100,
					ProfileID: profileID,
				},
			},
		},
		{
			name:       "err_invalid_id_format",
			id:         "not_number",
			wantStatus: http.StatusBadRequest,
			noUser:     true,
		},
		{
			name:       "err_no_user",
			id:         "100",
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_invalid_id",
			id:         "-100",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_random",
			id:         "100",
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.DeleteFavoriteQueryRequest{
					ID:        100,
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
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/userprofile/v1/queries/favorite/%s", tt.id), http.NoBody)
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Delete(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:        req,
				Handler:    api.serveDeleteFavoriteQuery,
				WantStatus: tt.wantStatus,
			})
		})
	}
}
