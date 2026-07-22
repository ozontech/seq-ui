package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetGroups(t *testing.T) {
	var (
		service       = "test-service"
		env           = "test-env"
		source        = "test-source"
		release       = "test-release"
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

		req     *errorgroups_v1.GetGroupsRequest
		want    *errorgroups_v1.GetGroupsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_duration",

			req: &errorgroups_v1.GetGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Duration:  durationpb.New(duration),
				Limit:     2,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_OLDEST,
				WithTotal: true,
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Total: 10,
				Groups: []*errorgroups_v1.GetGroupsResponse_Group{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
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
					Order:     types.OrderOldest,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_timerange",

			req: &errorgroups_v1.GetGroupsRequest{
				Service:  service,
				Env:      &env,
				Source:   &source,
				Release:  &release,
				Duration: durationpb.New(duration),
				TimeRange: &errorgroups_v1.TimeRange{
					From: timestamppb.New(twoMinutesAgo),
					To:   timestamppb.New(now),
				},
				Limit:     2,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_OLDEST,
				WithTotal: true,
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Total: 10,
				Groups: []*errorgroups_v1.GetGroupsResponse_Group{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
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
					Order:     types.OrderOldest,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_no_timerange",

			req: &errorgroups_v1.GetGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Limit:     2,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_OLDEST,
				WithTotal: true,
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Total: 10,
				Groups: []*errorgroups_v1.GetGroupsResponse_Group{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
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
					Order:     types.OrderOldest,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "ok_new",

			req: &errorgroups_v1.GetGroupsRequest{
				Service:   service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Limit:     2,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_OLDEST,
				WithTotal: true,
				Filter: &errorgroups_v1.GetGroupsRequest_Filter{
					IsNew: true,
				},
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Total: 10,
				Groups: []*errorgroups_v1.GetGroupsResponse_Group{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
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
					Order:     types.OrderOldest,
					WithTotal: true,
				},

				groups: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						Count:       10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						Count:       5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				total: 10,
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetGroupsRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{},

				err: someErr,
			},
		},
		{
			name: "err_svc_new",
			req: &errorgroups_v1.GetGroupsRequest{
				Filter: &errorgroups_v1.GetGroupsRequest_Filter{
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

			got, err := api.GetGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
