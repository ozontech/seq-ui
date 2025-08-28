package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetGroups(t *testing.T) {
	var (
		service       = "service1"
		env           = "prod"
		release       = "v1"
		duration      = 2 * time.Minute
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req        types.GetErrorGroupsRequest
		groupsResp []types.ErrorGroup
		countsResp uint64
		err        error
	}

	tests := []struct {
		name string

		req      *errorgroups_v1.GetGroupsRequest
		want     *errorgroups_v1.GetGroupsResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all_fields",
			req: &errorgroups_v1.GetGroupsRequest{
				Service:   service,
				Env:       &env,
				Release:   &release,
				Duration:  durationpb.New(2 * time.Minute),
				Limit:     10,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_OLDEST,
				WithTotal: true,
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Total: 2,
				Groups: []*errorgroups_v1.Group{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Env:       &env,
					Release:   &release,
					Duration:  &duration,
					Limit:     10,
					Offset:    0,
					Order:     types.OrderOldest,
					WithTotal: true,
				},
				groupsResp: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
				countsResp: 2,
			},
		},
		{
			name: "success_only_required_fields",
			req: &errorgroups_v1.GetGroupsRequest{
				Service: service,
			},
			want: &errorgroups_v1.GetGroupsResponse{
				Groups: []*errorgroups_v1.Group{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
					},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Limit:   25,
					Offset:  0,
					Order:   types.OrderFrequent,
				},
				groupsResp: []types.ErrorGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						SeenTotal:   10,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
					{
						Hash:        456,
						Message:     "some error 2",
						SeenTotal:   5,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
					},
				},
			},
		},
		{
			name: "err_no_service_field",
			req: &errorgroups_v1.GetGroupsRequest{
				Env:     &env,
				Release: &release,
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := repo_mock.NewMockRepository(ctrl)
			api := New(errorgroups.New(mockedRepo, config.LogTagsMapping{}))

			if tt.mockArgs != nil {
				mockedRepo.EXPECT().GetErrorGroups(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.groupsResp, tt.mockArgs.err).Times(1)
				if tt.mockArgs.req.WithTotal {
					mockedRepo.EXPECT().GetErrorGroupsCount(gomock.Any(), tt.mockArgs.req).
						Return(tt.mockArgs.countsResp, tt.mockArgs.err).Times(1)
				}
			}

			ctx := context.Background()
			got, err := api.GetGroups(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
