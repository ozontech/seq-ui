package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
)

func TestServeCancel(t *testing.T) {
	type mockArgs struct {
		err error
	}

	tests := []struct {
		name string

		req     cancelRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:     "ok",
			req:      cancelRequest{SessionID: sessionID},
			mockArgs: &mockArgs{},
		},
		{
			name:    "err_svc",
			req:     cancelRequest{SessionID: sessionID},
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
					CancelExport(gomock.Any(), tt.req.SessionID).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[cancelRequest, struct{}]{
				Method:  http.MethodPost,
				Target:  "/massexport/v1/cancel",
				Req:     tt.req,
				Handler: api.serveCancel,
				WantErr: tt.wantErr,
				NoResp:  true,
			})
		})
	}
}
