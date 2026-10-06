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

func TestGetServiceFilters(t *testing.T) {
	var (
		service = "test-service"
		env     = "test-env"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetServiceFiltersRequest

		filters []types.ServiceFilter
		err     error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetServiceFiltersRequest
		want    *errorgroups_v1.GetServiceFiltersResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetServiceFiltersRequest{
				Service: service,
				Env:     &env,
			},
			want: &errorgroups_v1.GetServiceFiltersResponse{
				Filters: []*errorgroups_v1.GetServiceFiltersResponse_Filter{
					{Key: "filter1", Values: []string{"val1", "val2"}},
					{Key: "filter2", Values: []string{"val10", "val20"}},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetServiceFiltersRequest{
					Service: service,
					Env:     &env,
				},

				filters: []types.ServiceFilter{
					{Key: "filter1", Values: []string{"val1", "val2"}},
					{Key: "filter2", Values: []string{"val10", "val20"}},
				},
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetServiceFiltersRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetServiceFiltersRequest{},

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
					GetServiceFilters(gomock.Any(), ma.req).
					Return(ma.filters, ma.err).
					Times(1)
			}

			got, err := api.GetServiceFilters(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
