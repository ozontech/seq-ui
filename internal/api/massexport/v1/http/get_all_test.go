package http

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestServeGetAll(t *testing.T) {
	type mockArgs struct {
		resp []types.ExportInfo
		err  error
	}

	tests := []struct {
		name string

		want    getAllResponse
		wantErr bool

		mockArgs mockArgs
	}{
		{
			name: "ok",
			mockArgs: mockArgs{
				resp: []types.ExportInfo{
					{
						ID:         sessionID,
						UserID:     userID,
						Status:     types.ExportStatusFinish,
						StartedAt:  from,
						FinishedAt: to,
					},
				},
			},
			want: getAllResponse{
				Exports: []checkResponse{
					{
						ID:         sessionID,
						UserID:     userID,
						Status:     exportStatusFinish,
						StartedAt:  from,
						FinishedAt: to,
						Duration:   (10 * time.Minute).String(),
					},
				},
			},
		},
		{
			name:    "err_svc",
			wantErr: true,
			mockArgs: mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, svcMock := setupAPI(t)

			svcMock.EXPECT().
				GetAll(gomock.Any()).
				Return(tt.mockArgs.resp, tt.mockArgs.err).
				Times(1)

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[struct{}, getAllResponse]{
				Method:  http.MethodGet,
				Target:  "/massexport/v1/jobs",
				Handler: api.serveJobs,
				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
