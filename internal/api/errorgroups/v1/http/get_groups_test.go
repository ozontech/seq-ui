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

func TestServeGetGroups(t *testing.T) {
	var (
		service       = "test-service"
		env           = "test-env"
		release       = "test-release"
		source        = "test-source"
		durationStr   = "2m"
		duration      = 2 * time.Minute
		now           = time.Now().Truncate(0).UTC()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorGroupsRequest

		groups []types.ErrorGroup
		total  uint64
		err    error
	}

	tests := []struct {
		name string

		req     getGroupsRequest
		want    getGroupsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_duration",

			req: getGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Duration:  &durationStr,
				Limit:     2,
				Offset:    0,
				Order:     OrderFrequent,
				WithTotal: true,
			},
			want: getGroupsResponse{
				Total: 10,
				Groups: []group{
					{
						Hash:        "123",
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        "456",
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Env:     &env,
					Source:  &source,
					Release: &release,
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
					Limit:     2,
					Offset:    0,
					Order:     types.OrderFrequent,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_timerange",

			req: getGroupsRequest{
				Service: service,
				Env:     &env,
				Source:  &source,
				Release: &release,
				TimeRange: &timeRange{
					From: twoMinutesAgo,
					To:   now,
				},
				Limit:     2,
				Offset:    0,
				Order:     OrderFrequent,
				WithTotal: true,
			},
			want: getGroupsResponse{
				Total: 10,
				Groups: []group{
					{
						Hash:        "123",
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        "456",
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Env:     &env,
					Source:  &source,
					Release: &release,
					TimeRange: &types.TimeRange{
						From: twoMinutesAgo,
						To:   now,
					},
					Limit:     2,
					Offset:    0,
					Order:     types.OrderFrequent,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_no_timerange",

			req: getGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Limit:     2,
				Offset:    0,
				Order:     OrderFrequent,
				WithTotal: true,
			},
			want: getGroupsResponse{
				Total: 10,
				Groups: []group{
					{
						Hash:        "123",
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        "456",
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
					Limit:     2,
					Offset:    0,
					Order:     types.OrderFrequent,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_new",

			req: getGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Limit:     2,
				Offset:    0,
				Order:     OrderFrequent,
				WithTotal: true,
				Filter: &groupsFilter{
					IsNew: true,
				},
			},
			want: getGroupsResponse{
				Total: 10,
				Groups: []group{
					{
						Hash:        "123",
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        "456",
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
					Limit:     2,
					Offset:    0,
					Order:     types.OrderFrequent,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "err_svc",

			req:     getGroupsRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{},

				err: someErr,
			},
		},
		{
			name: "err_svc_new",

			req: getGroupsRequest{
				Filter: &groupsFilter{
					IsNew: true,
				},
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{},

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
				if tt.req.Filter != nil && tt.req.Filter.IsNew {
					mockedSvc.EXPECT().
						GetNewErrorGroups(gomock.Any(), ma.req).
						Return(ma.groups, ma.total, ma.err).
						Times(1)
				} else {
					mockedSvc.EXPECT().
						GetErrorGroups(gomock.Any(), ma.req).
						Return(ma.groups, ma.total, ma.err).
						Times(1)
				}
			}

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getGroupsRequest, getGroupsResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/groups",
				Req:    tt.req,

				Handler: api.serveGetGroups,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
