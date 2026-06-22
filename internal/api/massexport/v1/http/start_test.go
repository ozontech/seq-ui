package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeStart(t *testing.T) {
	type mockArgs struct {
		resp types.StartExportResponse
		err  error
	}

	tests := []struct {
		name string

		req     startRequest
		want    startResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: startRequest{
				Query:  query,
				From:   from,
				To:     to,
				Window: window,
				Name:   testName,
			},
			want: startResponse{
				SessionID: sessionID,
			},
			mockArgs: &mockArgs{
				resp: types.StartExportResponse{
					SessionID: sessionID,
				},
			},
		},
		{
			name: "err_empty_name",
			req: startRequest{
				Query:  query,
				From:   from,
				To:     to,
				Window: window,
			},
			wantErr: true,
		},
		{
			name: "err_invalid_window",
			req: startRequest{
				Query:  query,
				From:   from,
				To:     to,
				Window: "invalid",
				Name:   testName,
			},
			wantErr: true,
		},
		{
			name: "err_window_larger_than_interval",
			req: startRequest{
				Query:  query,
				From:   from,
				To:     from.Add(5 * time.Second),
				Window: window,
				Name:   testName,
			},
			wantErr: true,
		},
		{
			name: "err_svc",
			req: startRequest{
				Query:  query,
				From:   from,
				To:     to,
				Window: window,
				Name:   testName,
			},
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
					StartExport(gomock.Any(), gomock.Any()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[startRequest, startResponse]{
				Method:  http.MethodPost,
				Target:  "/massexport/v1/start",
				Req:     tt.req,
				Handler: api.serveStart,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
