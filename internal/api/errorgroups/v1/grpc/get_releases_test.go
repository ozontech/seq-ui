package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repo_mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/ozontech/seq-ui/internal/pkg/service/errorgroups"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetReleases(t *testing.T) {
	var (
		service = "service1"
		env     = "prod"
	)

	type mockArgs struct {
		req  types.GetErrorGroupReleasesRequest
		resp []string
		err  error
	}

	tests := []struct {
		name string

		req      *errorgroups_v1.GetReleasesRequest
		want     *errorgroups_v1.GetReleasesResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all_fields",
			req: &errorgroups_v1.GetReleasesRequest{
				Service: service,
				Env:     &env,
			},
			want: &errorgroups_v1.GetReleasesResponse{
				Releases: []string{"v1", "v2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
					Env:     &env,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name: "success_no_env_field",
			req: &errorgroups_v1.GetReleasesRequest{
				Service: service,
			},
			want: &errorgroups_v1.GetReleasesResponse{
				Releases: []string{"v1", "v2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name: "success_no_group_hash_field",
			req: &errorgroups_v1.GetReleasesRequest{
				Service: service,
				Env:     &env,
			},
			want: &errorgroups_v1.GetReleasesResponse{
				Releases: []string{"v1", "v2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
					Env:     &env,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name: "success_only_service_field",
			req: &errorgroups_v1.GetReleasesRequest{
				Service: service,
			},
			want: &errorgroups_v1.GetReleasesResponse{
				Releases: []string{"v1", "v2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetErrorGroupReleasesRequest{
					Service: service,
				},
				resp: []string{"v1", "v2"},
			},
		},
		{
			name: "err_no_service_field",
			req: &errorgroups_v1.GetReleasesRequest{
				Env: &env,
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
				mockedRepo.EXPECT().GetErrorReleases(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			got, err := api.GetReleases(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
