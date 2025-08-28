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

func TestServeCreate(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "my_dashboard"
	dashboardMeta := "my_meta"

	formatReqBody := func(name, meta string) string {
		return fmt.Sprintf(`{"name":%q,"meta":%q}`, name, meta)
	}

	type mockArgs struct {
		req  types.CreateDashboardRequest
		resp string
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
			reqBody:      formatReqBody(dashboardName, dashboardMeta),
			wantRespBody: fmt.Sprintf(`{"uuid":%q}`, dashboardUUID),
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					ProfileID: profileID,
					Name:      dashboardName,
					Meta:      dashboardMeta,
				},
				resp: dashboardUUID,
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
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_empty_name",
			reqBody:    formatReqBody("", dashboardMeta),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_empty_meta",
			reqBody:    formatReqBody(dashboardName, ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_random",
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					ProfileID: profileID,
					Name:      dashboardName,
					Meta:      dashboardMeta,
				},
				err: errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newTestData(t)
			req := httptest.NewRequest(http.MethodPost, "/dashboards/v1/", strings.NewReader(tt.reqBody))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Create(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveCreate,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
