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

func TestServeUpdate(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "my_dashboard"
	dashboardMeta := "my_meta"

	formatReqBody := func(name, meta string) string {
		var sb strings.Builder
		sb.WriteString("{")
		if name != "" {
			sb.WriteString(fmt.Sprintf(`"name":%q`, name))
		}
		if meta != "" {
			if name != "" {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf(`"meta":%q`, meta))
		}
		sb.WriteString("}")
		return sb.String()
	}

	type mockArgs struct {
		req types.UpdateDashboardRequest
		err error
	}

	tests := []struct {
		name string

		uuid       string
		reqBody    string
		wantStatus int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:       "success_all",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
			},
		},
		{
			name:       "success_only_name",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody(dashboardName, ""),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
				},
			},
		},
		{
			name:       "success_only_meta",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody("", dashboardMeta),
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Meta:      &dashboardMeta,
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
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_invalid_uuid",
			uuid:       "invalid-uuid",
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_svc_empty_request",
			uuid:       dashboardUUID,
			reqBody:    `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_not_found",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusNotFound,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
				err: types.ErrNotFound,
			},
		},
		{
			name:       "err_repo_permission_denied",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusForbidden,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
				},
				err: types.ErrPermissionDenied,
			},
		},
		{
			name:       "err_repo_random",
			uuid:       dashboardUUID,
			reqBody:    formatReqBody(dashboardName, dashboardMeta),
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					ProfileID: profileID,
					UUID:      dashboardUUID,
					Name:      &dashboardName,
					Meta:      &dashboardMeta,
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
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/dashboards/v1/%s", tt.uuid), strings.NewReader(tt.reqBody))
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().Update(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:        req,
				Handler:    api.serveUpdate,
				WantStatus: tt.wantStatus,
			})
		})
	}
}
