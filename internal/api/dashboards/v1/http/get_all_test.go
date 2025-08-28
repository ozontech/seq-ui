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

func TestServeGetAll(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	limit := 2
	offset := 0

	formatReqBody := func(limit, offset int) string {
		return fmt.Sprintf(`{"limit":%d,"offset":%d}`, limit, offset)
	}

	type mockArgs struct {
		req  types.GetAllDashboardsRequest
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
			reqBody:      formatReqBody(limit, offset),
			wantRespBody: `{"dashboards":[{"uuid":"064dc707-02b8-7000-8201-02a7f396738a","name":"dashboard1","owner_name":"user1"},{"uuid":"064dc707-12b9-7000-a238-682b044c908b","name":"dashboard2","owner_name":"user2"}]}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.GetAllDashboardsRequest{
					Limit:  limit,
					Offset: offset,
				},
				resp: types.DashboardInfosWithOwner{
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-02b8-7000-8201-02a7f396738a",
							Name: "dashboard1",
						},
						OwnerName: "user1",
					},
					{
						DashboardInfo: types.DashboardInfo{
							UUID: "064dc707-12b9-7000-a238-682b044c908b",
							Name: "dashboard2",
						},
						OwnerName: "user2",
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
			reqBody:    formatReqBody(limit, offset),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_invalid_limit",
			reqBody:    formatReqBody(0, offset),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_invalid_offset",
			reqBody:    formatReqBody(limit, -10),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_random",
			reqBody:    formatReqBody(limit, offset),
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.GetAllDashboardsRequest{
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

			req := httptest.NewRequest(http.MethodPost, "/dashboards/v1/all", strings.NewReader(tt.reqBody))
			api, mockedRepo := newTestData(t)

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
				Handler:      api.serveGetAll,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
