package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/dashboards/v1"
)

func TestGetByUUID(t *testing.T) {
	type mockArgs struct {
		uuid string
		resp types.Dashboard
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.GetByUUIDRequest
		want     *dashboards.GetByUUIDResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &dashboards.GetByUUIDRequest{
				Uuid: dashboardUUID,
			},
			want: &dashboards.GetByUUIDResponse{
				Name:      dashboardName,
				Meta:      dashboardMeta,
				OwnerName: dashboardOwner,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				resp: types.Dashboard{
					Name:      dashboardName,
					Meta:      dashboardMeta,
					OwnerName: dashboardOwner,
				},
			},
		},
		{
			name: "err_svc",
			req: &dashboards.GetByUUIDRequest{
				Uuid: dashboardUUID,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				uuid: dashboardUUID,
				err:  errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					GetDashboardByUUID(gomock.Any(), tt.mockArgs.uuid).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.GetByUUID(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
