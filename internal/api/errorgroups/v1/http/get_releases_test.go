package http

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
)

func TestServeGetReleases(t *testing.T) {
	var (
		service = "test-service"
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetReleasesRequest

		releases []string
		err      error
	}

	tests := []struct {
		name string

		req     getReleasesRequest
		want    getReleasesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: getReleasesRequest{
				Service: service,
				Env:     &env,
			},
			want: getReleasesResponse{
				Releases: []string{"release1", "release2"},
			},

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{
					Service: service,
					Env:     &env,
				},

				releases: []string{"release1", "release2"},
			},
		},
		{
			name: "err_svc",

			req:     getReleasesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedSvc := svc_mock.NewMockService(ctrl)

			api := New(mockedSvc)

			if ma := tt.mockArgs; ma != nil {
				mockedSvc.EXPECT().
					GetReleases(gomock.Any(), ma.req).
					Return(ma.releases, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getReleasesRequest, getReleasesResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/releases",
				Req:    tt.req,

				Handler: api.serveGetReleases,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
