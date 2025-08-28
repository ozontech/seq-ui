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

func TestGetServices(t *testing.T) {
	var (
		query = "itc"
		env   = "prod"
	)

	type mockArgs struct {
		req  types.GetServicesRequest
		resp []string
		err  error
	}

	tests := []struct {
		name string

		req      *errorgroups_v1.GetServicesRequest
		want     *errorgroups_v1.GetServicesResponse
		wantCode codes.Code

		mockArgs *mockArgs
		noUser   bool
	}{
		{
			name: "success_all_fields",
			req: &errorgroups_v1.GetServicesRequest{
				Query: query,
				Env:   &env,
			},
			want: &errorgroups_v1.GetServicesResponse{
				Services: []string{"service1", "service2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query: query,
					Env:   &env,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name: "success_no_env_field",
			req: &errorgroups_v1.GetServicesRequest{
				Query: query,
			},
			want: &errorgroups_v1.GetServicesResponse{
				Services: []string{"service1", "service2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query: query,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name: "success_no_query_field",
			req: &errorgroups_v1.GetServicesRequest{
				Env: &env,
			},
			want: &errorgroups_v1.GetServicesResponse{
				Services: []string{"service1", "service2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Env: &env,
				},
				resp: []string{"service1", "service2"},
			},
		},
		{
			name: "success_no_query_no_env",
			req:  &errorgroups_v1.GetServicesRequest{},
			want: &errorgroups_v1.GetServicesResponse{
				Services: []string{"service1", "service2"},
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req:  types.GetServicesRequest{},
				resp: []string{"service1", "service2"},
			},
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
				mockedRepo.EXPECT().GetServices(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).Times(1)
			}

			ctx := context.Background()
			got, err := api.GetServices(ctx, tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
