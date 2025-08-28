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

func TestServeGetByUUID(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"

	type mockArgs struct {
		uuid string
		resp types.Dashboard
		err  error
	}

	tests := []struct {
		name string

		uuid         string
		wantRespBody string
		wantStatus   int

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name:         "success",
			uuid:         dashboardUUID,
			wantRespBody: `{"name":"dashboard1","meta":"meta1","owner_name":"owner"}`,
			wantStatus:   http.StatusOK,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				resp: types.Dashboard{
					OwnerName: "owner",
					Name:      "dashboard1",
					Meta:      "meta1",
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
			name:       "err_repo_not_found",
			uuid:       dashboardUUID,
			wantStatus: http.StatusNotFound,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  types.ErrNotFound,
			},
		},
		{
			name:       "err_repo_random",
			uuid:       dashboardUUID,
			wantStatus: http.StatusInternalServerError,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  errors.New("random repo err"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedRepo := newTestData(t)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/dashboards/v1/%s", tt.uuid), http.NoBody)
			rCtx := chi.NewRouteContext()
			rCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rCtx))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetByUUID(gomock.Any(), tt.mockArgs.uuid).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}
			if !tt.noUser {
				req = req.WithContext(context.WithValue(req.Context(), types.UserKey{}, userName))
				api.profiles.SetID(userName, profileID)
			}

			httputil.DoTestHTTP(t, httputil.TestDataHTTP{
				Req:          req,
				Handler:      api.serveGetByUUID,
				WantRespBody: tt.wantRespBody,
				WantStatus:   tt.wantStatus,
			})
		})
	}
}
