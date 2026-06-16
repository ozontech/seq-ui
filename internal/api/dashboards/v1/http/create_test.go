package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/service/dashboards/mock"
)

func setupAPI(t *testing.T) (*API, *mock.MockService) {
	ctrl := gomock.NewController(t)
	mockedSvc := mock.NewMockService(ctrl)
	p := profiles.New(mockedSvc)
	return New(mockedSvc, p), mockedSvc
}

func TestServeCreate(t *testing.T) {
	userName := "unnamed"
	var profileID int64 = 1
	dashboardUUID := "064dc707-02b8-7000-8201-02a7f396738a"
	dashboardName := "my_dashboard"
	dashboardMeta := "my_meta"

	type mockArgs struct {
		req  types.CreateDashboardRequest
		resp string
		err  error
	}

	tests := []struct {
		name string

		req     createRequest
		want    createResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  createRequest{Name: dashboardName, Meta: dashboardMeta},
			want: createResponse{UUID: dashboardUUID},
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
			name:    "err_svc",
			req:     createRequest{Name: dashboardName, Meta: dashboardMeta},
			wantErr: true,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					ProfileID: profileID,
					Name:      dashboardName,
					Meta:      dashboardMeta,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
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
