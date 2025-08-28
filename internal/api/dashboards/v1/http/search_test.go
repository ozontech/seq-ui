package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"go.uber.org/mock/gomock"
)

func TestServeSearch(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	query := "test"
	limit := 2
	offset := 0
	filter := &types.SearchDashboardsFilter{
		OwnerName: &userName,
	}

	formatReqBody := func(query string, limit, offset int, filter *types.SearchDashboardsFilter) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(`{"query":%q,"limit":%d,"offset":%d`, query, limit, offset))
		if filter != nil {
			sb.WriteString(`,"filter":{`)
			if filter.OwnerName != nil {
				sb.WriteString(fmt.Sprintf(`"owner_name":%q`, *filter.OwnerName))
			}
			sb.WriteString("}")
		}
		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req  types.SearchDashboardsRequest
		resp types.DashboardInfosWithOwner
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
			reqBody:      formatReqBody(query, limit, offset, nil),
			wantRespBody: `{"dashboards":[{"uuid":"064dc707-02b8-7000-8201-02a7f396738a","name":"my test dashboard","owner_name":"user1"},{"uuid":"064dc707-12b9-7000-a238-682b044c908b","name":"tested","owner_name":"user2"}]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
					Limit:  limit,
					Offset: offset,
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "my test dashboard",
						},
						OwnerName: "user1",
					},
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-12b9-7000-a238-682b044c908b",
							Name: "tested",
						},
						OwnerName: "user2",
					},
				},
			},
		},
		{
			name:         "success_with_filter",
			reqBody:      formatReqBody(query, limit, offset, filter),
			wantRespBody: fmt.Sprintf(`{"dashboards":[{"uuid":"064dc707-02b8-7000-8201-02a7f396738a","name":"my test dashboard","owner_name":%q}]}`, userName),
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
					Limit:  limit,
					Offset: offset,
					Filter: filter,
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "my test dashboard",
						},
						OwnerName: userName,
					},
				},
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
			reqBody:    formatReqBody(query, limit, offset, nil),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_invalid_limit",
			reqBody:    formatReqBody(query, 0, offset, nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_invalid_offset",
			reqBody:    formatReqBody(query, limit, -10, nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_random",
			reqBody:    formatReqBody(query, limit, offset, nil),
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.SearchDashboardsRequest{
					Query:  query,
					Limit:  limit,
					Offset: offset,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/dashboards/v1/search", strings.NewReader(tt.reqBody))
			api, mockedRepo := newTestData(t)

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Search(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveSearch,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
