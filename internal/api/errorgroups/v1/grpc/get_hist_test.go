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

func TestGetHist(t *testing.T) {
	var (
		service              = "service1"
		groupHash     uint64 = 123
		env                  = "prod"
		release              = "v1"
		twoMinutes           = 2 * time.Minute
		now                  = time.Now()
		oneMinuteAgo         = now.Add(-1 * time.Minute)
		twoMinutesAgo        = now.Add(-2 * time.Minute)
	)

	type mockArgs struct {
		req  types.GetErrorHistRequest
		resp []types.ErrorHistBucket
		err  error
	}

	tests := []struct {
		name string

		req      *errorgroups_v1.GetHistRequest
		want     *errorgroups_v1.GetHistResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all_fields",
			req: &errorgroups_v1.GetHistRequest{
				Service:   service,
				GroupHash: &groupHash,
				Env:       &env,
				Release:   &release,
				Duration:  durationpb.New(2 * time.Minute),
			},
			want: &errorgroups_v1.GetHistResponse{
				Buckets: []*errorgroups_v1.Bucket{
					{Time: timestamppb.New(oneMinuteAgo), Count: 10},
					{Time: timestamppb.New(twoMinutesAgo), Count: 20},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					Service:   service,
					GroupHash: &groupHash,
					Env:       &env,
					Release:   &release,
					Duration:  &twoMinutes,
				},
				resp: []types.ErrorHistBucket{
					{Time: oneMinuteAgo, Count: 10},
					{Time: twoMinutesAgo, Count: 20},
				},
			},
		},
		{
			name: "success_only_service_field",
			req: &errorgroups_v1.GetHistRequest{
				Service: service,
			},
			want: &errorgroups_v1.GetHistResponse{
				Buckets: []*errorgroups_v1.Bucket{
					{Time: timestamppb.New(oneMinuteAgo), Count: 10},
					{Time: timestamppb.New(twoMinutesAgo), Count: 20},
				},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					Service: service,
				},
				resp: []types.ErrorHistBucket{
					{Time: oneMinuteAgo, Count: 10},
					{Time: twoMinutesAgo, Count: 20},
				},
			},
		},
		{
			name: "err_no_service_field",
			req: &errorgroups_v1.GetHistRequest{
				GroupHash: &groupHash,
				Env:       &env,
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
				mockedRepo.EXPECT().GetErrorHist(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			got, err := api.GetHist(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
