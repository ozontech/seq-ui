package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetReleases(t *testing.T) {
	var (
		service = "test-service"
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetReleasesRequest

		releases []string
		err      error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetReleasesRequest
		want    *errorgroups_v1.GetReleasesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetReleasesRequest{
				Service: service,
				Env:     &env,
			},
			want: &errorgroups_v1.GetReleasesResponse{
				Releases: []string{"release1", "release2"},
			},

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{
					Service: service,
					Env:     &env,
				},

				releases: []string{"release1", "release2"},
			},
		},
		{
			name:    "err_svc",
			req:     &errorgroups_v1.GetReleasesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{},

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
					GetReleases(gomock.Any(), ma.req).
					Return(ma.releases, ma.err).
					Times(1)
			}

			got, err := api.GetReleases(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
