package http

import (
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
)

func TestServeRestore(t *testing.T) {
	type mockArgs struct {
		err error
	}

	tests := []struct {
		name string

		req     restoreRequest
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name:     "ok",
			req:      restoreRequest{SessionID: sessionID},
			mockArgs: &mockArgs{},
		},
		{
			name:    "err_svc",
			req:     restoreRequest{SessionID: sessionID},
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
					RestoreExport(gomock.Any(), tt.req.SessionID).
					Return(tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[restoreRequest, struct{}]{
				Method:  http.MethodPost,
				Target:  "/massexport/v1/restore",
				Req:     tt.req,
				Handler: api.serveRestore,
				WantErr: tt.wantErr,
				NoResp:  true,
			})
		})
	}
}
