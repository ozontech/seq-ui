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

func TestCreate(t *testing.T) {
	type mockArgs struct {
		req  types.CreateDashboardRequest
		resp string
		err  error
	}

	tests := []struct {
		name string

		req      *dashboards.CreateRequest
		want     *dashboards.CreateResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &dashboards.CreateRequest{
				Name: testDashboardName,
				Meta: testDashboardMeta,
			},
			want: &dashboards.CreateResponse{
				Uuid: testDashboardUUID,
			},
			wantCode: codes.OK,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					Name: testDashboardName,
					Meta: testDashboardMeta,
				},
				resp: testDashboardUUID,
			},
		},
		{
			name: "err_svc",
			req: &dashboards.CreateRequest{
				Name: testDashboardName,
				Meta: testDashboardMeta,
			},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				req: types.CreateDashboardRequest{
					Name: testDashboardName,
					Meta: testDashboardMeta,
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
					CreateDashboard(gomock.Any(), tt.mockArgs.req).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.Create(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
