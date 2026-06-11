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

func TestGetServices(t *testing.T) {
	var (
		query   = "test-query"
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetServicesRequest

		services []string
		err      error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetServicesRequest
		want    *errorgroups_v1.GetServicesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetServicesRequest{
				Query:  query,
				Env:    &env,
				Limit:  10,
				Offset: 20,
			},
			want: &errorgroups_v1.GetServicesResponse{
				Services: []string{"service1", "service2"},
			},

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{
					Query:  query,
					Env:    &env,
					Limit:  10,
					Offset: 20,
				},

				services: []string{"service1", "service2"},
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetServicesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{},

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
				mockedSvc.EXPECT().
					GetServices(gomock.Any(), ma.req).
					Return(ma.services, ma.err).
					Times(1)
			}

			got, err := api.GetServices(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
