package http

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
)

func TestServeDiffByReleases(t *testing.T) {
	var (
		service       = "test-service"
		releases      = []string{"release1", "release2"}
		env           = "test-env"
		source        = "test-source"
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.DiffByReleasesRequest

		groups []types.DiffGroup
		total  uint64
		err    error
	}

	tests := []struct {
		name string

		req     diffByReleasesRequest
		want    diffByReleasesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: diffByReleasesRequest{
				Service:   service,
				Releases:  releases,
				Env:       &env,
				Source:    &source,
				Limit:     2,
				Offset:    0,
				Order:     OrderFrequent,
				WithTotal: true,
			},
			want: diffByReleasesResponse{
				Total: 10,
				Groups: []diffGroup{
					{
						Hash:        "123",
						Message:     "some error 1",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]diffReleaseInfo{
							"release1": {SeenTotal: 10},
							"release2": {SeenTotal: 20},
						},
					},
					{
						Hash:        "456",
						Message:     "some error 2",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]diffReleaseInfo{
							"release1": {SeenTotal: 40},
							"release2": {SeenTotal: 0},
						},
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:   service,
					Releases:  releases,
					Env:       &env,
					Source:    &source,
					Limit:     2,
					Offset:    0,
					Order:     types.OrderFrequent,
					WithTotal: true,
				},

				groups: []types.DiffGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]types.DiffReleaseInfo{
							"release1": {SeenTotal: 10},
							"release2": {SeenTotal: 20},
						},
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]types.DiffReleaseInfo{
							"release1": {SeenTotal: 40},
							"release2": {SeenTotal: 0},
						},
					},
				},
				total: 10,
			},
		},

		{
			name: "err_svc",

			req:     diffByReleasesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedSvc := svc_mock.NewMockService(ctrl)

			api := New(mockedSvc)

			if ma := tt.mockArgs; ma != nil {
				mockedSvc.EXPECT().
					DiffByReleases(gomock.Any(), ma.req).
					Return(ma.groups, ma.total, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[diffByReleasesRequest, diffByReleasesResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/diff_by_releases",
				Req:    tt.req,

				Handler: api.serveDiffByReleases,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
