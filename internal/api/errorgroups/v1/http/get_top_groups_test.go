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

func TestServeGetTopGroups(t *testing.T) {
	var (
		env           = "test-env"
		source        = "test-source"
		durationStr   = "2m"
		duration      = 2 * time.Minute
		now           = time.Now().Truncate(0)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetTopErrorGroupsRequest

		groups []types.TopErrorGroup
		total  uint64
		err    error
	}

	tests := []struct {
		name string

		req     getTopGroupsRequest
		want    getTopGroupsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_duration",

			req: getTopGroupsRequest{
				Env:       &env,
				Source:    &source,
				Duration:  &durationStr,
				Limit:     2,
				Offset:    0,
				WithTotal: true,
			},
			want: getTopGroupsResponse{
				Total: 10,
				Groups: []topGroup{
					{
						Hash:      "123",
						Message:   "some error 1",
						Source:    source,
						SeenTotal: 5,
					},
					{
						Hash:      "456",
						Message:   "some error 2",
						Source:    source,
						SeenTotal: 10,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Env:    &env,
					Source: &source,
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
					Limit:     2,
					Offset:    0,
					WithTotal: true,
				},

				groups: []types.TopErrorGroup{
					{
						Hash:    123,
						Message: "some error 1",
						Source:  source,
						Count:   5,
					},
					{
						Hash:    456,
						Message: "some error 2",
						Source:  source,
						Count:   10,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_timerange",

			req: getTopGroupsRequest{
				Env:    &env,
				Source: &source,
				TimeRange: &timeRange{
					From: twoMinutesAgo,
					To:   now,
				},
				Limit:     2,
				Offset:    0,
				WithTotal: true,
			},
			want: getTopGroupsResponse{
				Total: 10,
				Groups: []topGroup{
					{
						Hash:      "123",
						Message:   "some error 1",
						Source:    source,
						SeenTotal: 5,
					},
					{
						Hash:      "456",
						Message:   "some error 2",
						Source:    source,
						SeenTotal: 10,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Env:    &env,
					Source: &source,
					TimeRange: &types.TimeRange{
						From: twoMinutesAgo,
						To:   now,
					},
					Limit:     2,
					Offset:    0,
					WithTotal: true,
				},

				groups: []types.TopErrorGroup{
					{
						Hash:    123,
						Message: "some error 1",
						Source:  source,
						Count:   5,
					},
					{
						Hash:    456,
						Message: "some error 2",
						Source:  source,
						Count:   10,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_no_timerange",

			req: getTopGroupsRequest{
				Env:       &env,
				Source:    &source,
				Limit:     2,
				Offset:    0,
				WithTotal: true,
			},
			want: getTopGroupsResponse{
				Total: 10,
				Groups: []topGroup{
					{
						Hash:      "123",
						Message:   "some error 1",
						Source:    source,
						SeenTotal: 5,
					},
					{
						Hash:      "456",
						Message:   "some error 2",
						Source:    source,
						SeenTotal: 10,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Env:       &env,
					Source:    &source,
					Limit:     2,
					Offset:    0,
					WithTotal: true,
				},

				groups: []types.TopErrorGroup{
					{
						Hash:    123,
						Message: "some error 1",
						Source:  source,
						Count:   5,
					},
					{
						Hash:    456,
						Message: "some error 2",
						Source:  source,
						Count:   10,
					},
				},
				total: 10,
			},
		},
		{
			name: "err_svc",

			req:     getTopGroupsRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{},

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
					GetTopErrorGroups(gomock.Any(), ma.req).
					Return(ma.groups, ma.total, ma.err).
					Times(1)
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getTopGroupsRequest, getTopGroupsResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/top_groups",
				Req:    tt.req,

				Handler: api.serveGetTopGroups,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
