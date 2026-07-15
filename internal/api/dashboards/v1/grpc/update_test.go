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

func TestUpdate(t *testing.T) {
	type mockArgs struct {
		req types.UpdateDashboardRequest
		err error
	}

	tests := []struct {
		name string

		req      *dashboards.UpdateRequest
		want     *dashboards.UpdateResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &dashboards.UpdateRequest{
				Uuid: testDashboardUUID,
				Name: &testDashboardName,
				Meta: &testDashboardMeta,
			},
			want:     &dashboards.UpdateResponse{},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					UUID: testDashboardUUID,
					Name: &testDashboardName,
					Meta: &testDashboardMeta,
				},
			},
		},
		{
			name: "err_svc",
			req: &dashboards.UpdateRequest{
				Uuid: testDashboardUUID,
				Name: &testDashboardName,
				Meta: &testDashboardMeta,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.UpdateDashboardRequest{
					UUID: testDashboardUUID,
					Name: &testDashboardName,
					Meta: &testDashboardMeta,
				},
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, mockedSvc := setupTestAPI(t)

			if tt.mockArgs != nil {
				mockedSvc.EXPECT().
					UpdateDashboard(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.err).
					Times(1)
			}

			got, err := api.Update(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
