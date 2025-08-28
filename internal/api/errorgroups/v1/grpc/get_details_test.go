package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetDetails(t *testing.T) {
	var (
		service              = "service1"
		groupHash     uint64 = 123
		env                  = "prod"
		release              = "test_release"
		now                  = time.Now()
		oneMinuteAgo         = now.Add(-1 * time.Minute)
		twoMinutesAgo        = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req         types.GetErrorGroupDetailsRequest
		detailsResp types.ErrorGroupDetails
		countsResp  *types.ErrorGroupCounts
		err         error
	}

	tests := []struct {
		name string

		req      *errorgroups_v1.GetDetailsRequest
		want     *errorgroups_v1.GetDetailsResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "success_all_fields",
			req: &errorgroups_v1.GetDetailsRequest{
				Service:   service,
				GroupHash: groupHash,
				Env:       &env,
				Release:   &release,
			},
			want: &errorgroups_v1.GetDetailsResponse{
				GroupHash:   123,
				Message:     "some error",
				SeenTotal:   10,
				FirstSeenAt: timestamppb.New(twoMinutesAgo),
				LastSeenAt:  timestamppb.New(oneMinuteAgo),
				Distributions: &errorgroups_v1.GetDetailsResponse_Distributions{
					ByEnv: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: env, Percent: 100},
					},
					ByRelease: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: release, Percent: 100},
					},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					Service:   service,
					GroupHash: groupHash,
					Env:       &env,
					Release:   &release,
				},
				detailsResp: types.ErrorGroupDetails{
					GroupHash:   123,
					Message:     "some error",
					SeenTotal:   10,
					FirstSeenAt: twoMinutesAgo,
					LastSeenAt:  oneMinuteAgo,
				},
			},
		},
		{
			name: "success_only_required_fields",
			req: &errorgroups_v1.GetDetailsRequest{
				Service:   service,
				GroupHash: groupHash,
			},
			want: &errorgroups_v1.GetDetailsResponse{
				GroupHash:   123,
				Message:     "some error",
				SeenTotal:   10,
				FirstSeenAt: timestamppb.New(twoMinutesAgo),
				LastSeenAt:  timestamppb.New(oneMinuteAgo),
				Distributions: &errorgroups_v1.GetDetailsResponse_Distributions{
					ByEnv: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "env1", Percent: 70},
						{Value: "env2", Percent: 30},
					},
					ByRelease: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "release2", Percent: 60},
						{Value: "release1", Percent: 40},
					},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					Service:   service,
					GroupHash: groupHash,
				},
				detailsResp: types.ErrorGroupDetails{
					GroupHash:   123,
					Message:     "some error",
					SeenTotal:   10,
					FirstSeenAt: twoMinutesAgo,
					LastSeenAt:  oneMinuteAgo,
				},
				countsResp: &types.ErrorGroupCounts{
					ByEnv: types.ErrorGroupCount{
						"env1": 7,
						"env2": 3,
					},
					ByRelease: types.ErrorGroupCount{
						"release1": 4,
						"release2": 6,
					},
				},
			},
		},
		{
			name: "err_no_service_field",
			req: &errorgroups_v1.GetDetailsRequest{
				GroupHash: groupHash,
				Env:       &env,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_no_group_hash_field",
			req: &errorgroups_v1.GetDetailsRequest{
				Service: service,
				Env:     &env,
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
				mockedRepo.EXPECT().GetErrorDetails(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.detailsResp, tt.mockArgs.err).Times(1)
				if tt.mockArgs.countsResp != nil {
					mockedRepo.EXPECT().GetErrorCounts(gomock.Any(), tt.mockArgs.req).
						Return(*tt.mockArgs.countsResp, tt.mockArgs.err).Times(1)
				}
			}

			ctx := context.Background()
			got, err := api.GetDetails(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
