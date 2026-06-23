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

func TestServeGetHist(t *testing.T) {
	var (
		groupHashStr  = "123"
		groupHash     = uint64(123)
		service       = "test-service"
		env           = "test-env"
		source        = "test-source"
		release       = "test-release"
		durationStr   = "2m"
		duration      = 2 * time.Minute
		now           = time.Now().Truncate(0)
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorHistRequest

		hist types.ErrorHist
		err  error
	}

	tests := []struct {
		name string

		req     getHistRequest
		want    getHistResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_duration",

			req: getHistRequest{
				GroupHash: &groupHashStr,
				Service:   &service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Duration:  &durationStr,
			},
			want: getHistResponse{
				Buckets: []bucket{
					{
						Time:  twoMinutesAgo,
						Count: 100,
					},
					{
						Time:  oneMinuteAgo,
						Count: 200,
					},
				},
				Interval: 123,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					GroupHash: &groupHash,
					Service:   &service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
				},

				hist: types.ErrorHist{
					Buckets: []types.ErrorHistBucket{
						{Time: twoMinutesAgo, Count: 100},
						{Time: oneMinuteAgo, Count: 200},
					},
					Interval: 123,
				},
			},
		},
		{
			name: "ok_timerange",

			req: getHistRequest{
				GroupHash: &groupHashStr,
				Service:   &service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				TimeRange: &timeRange{
					From: twoMinutesAgo,
					To:   now,
				},
			},
			want: getHistResponse{
				Buckets: []bucket{
					{
						Time:  twoMinutesAgo,
						Count: 100,
					},
					{
						Time:  oneMinuteAgo,
						Count: 200,
					},
				},
				Interval: 123,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					GroupHash: &groupHash,
					Service:   &service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
					TimeRange: &types.TimeRange{
						From: twoMinutesAgo,
						To:   now,
					},
				},

				hist: types.ErrorHist{
					Buckets: []types.ErrorHistBucket{
						{Time: twoMinutesAgo, Count: 100},
						{Time: oneMinuteAgo, Count: 200},
					},
					Interval: 123,
				},
			},
		},
		{
			name: "ok_no_timerange",

			req: getHistRequest{
				GroupHash: &groupHashStr,
				Service:   &service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
			},
			want: getHistResponse{
				Buckets: []bucket{
					{
						Time:  twoMinutesAgo,
						Count: 100,
					},
					{
						Time:  oneMinuteAgo,
						Count: 200,
					},
				},
				Interval: 123,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					GroupHash: &groupHash,
					Service:   &service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
				},

				hist: types.ErrorHist{
					Buckets: []types.ErrorHistBucket{
						{Time: twoMinutesAgo, Count: 100},
						{Time: oneMinuteAgo, Count: 200},
					},
					Interval: 123,
				},
			},
		},
		{
			name: "err_svc",

			req:     getHistRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{},

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
					GetHist(gomock.Any(), ma.req).
					Return(ma.hist, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getHistRequest, getHistResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/hist",
				Req:    tt.req,

				Handler: api.serveGetHist,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
