package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"go.uber.org/mock/gomock"
)

func TestServeDelete(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"

	type mockArgs struct {
		req types.DeleteDashboardRequest
		err error
	}

	tests := []struct {
		name string

		uuid       string
		wantStatus int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:       "success",
			uuid:       dashboardUUID,
			wantStatus: http.StatusOK,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
					ProfileID: profileID,
				},
			},
		},
		{
			name:       "err_no_user",
			wantStatus: http.StatusUnauthorized,
			noUser:     true,
		},
		{
			name:       "err_svc_invalid_uuid",
			uuid:       "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "err_repo_permission_denied",
			uuid:       dashboardUUID,
			wantStatus: http.StatusForbidden,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
					ProfileID: profileID,
				},
				err: types.ErrPermissionDenied,
			},
		},
		{
			name:       "err_repo_random",
			uuid:       dashboardUUID,
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				req: types.DeleteDashboardRequest{
					UUID:      dashboardUUID,
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

			api, mockedRepo := newTestData(t)
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/dashboards/v1/%s", tt.uuid), http.NoBody)
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("uuid", tt.uuid)
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
				Handler:    api.serveDelete,
				WantStatus: tt.wantStatus,
			})
		})
	}
}
