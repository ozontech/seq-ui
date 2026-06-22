package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeCheck(t *testing.T) {
	type mockArgs struct {
		resp types.ExportInfo
		err  error
	}

	tests := []struct {
		name string

		req     checkRequest
		want    checkResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  checkRequest{SessionID: sessionID},
			mockArgs: &mockArgs{
				resp: types.ExportInfo{
					ID:         sessionID,
					UserID:     userID,
					Status:     types.ExportStatusFinish,
					StartedAt:  from,
					FinishedAt: to,
				},
			},
			want: checkResponse{
				ID:         sessionID,
				UserID:     userID,
				Status:     exportStatusFinish,
				StartedAt:  from,
				FinishedAt: to,
				Duration:   (10 * time.Minute).String(),
			},
		},
		{
			name:    "err_svc",
			req:     checkRequest{SessionID: sessionID},
			wantErr: true,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, svcMock := setupAPI(t)

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					CheckExport(gomock.Any(), tt.req.SessionID).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[checkRequest, checkResponse]{
				Method:  http.MethodPost,
				Target:  "/massexport/v1/check",
				Req:     tt.req,
				Handler: api.serveCheck,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
