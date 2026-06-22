package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

func TestRestore(t *testing.T) {
	type mockArgs struct {
		err error
	}

	tests := []struct {
		name string

		req      *massexport_v1.RestoreRequest
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name:     "ok",
			req:      &massexport_v1.RestoreRequest{SessionId: sessionID},
			mockArgs: &mockArgs{},
		},
		{
			name:     "err_svc",
			req:      &massexport_v1.RestoreRequest{SessionId: sessionID},
			wantCode: codes.Internal,
			mockArgs: &mockArgs{
				err: errSomethingWrong,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api, svcMock := setupAPI(t)

			if tt.mockArgs != nil {
				svcMock.EXPECT().
					RestoreExport(gomock.Any(), tt.req.GetSessionId()).
					Return(tt.mockArgs.err).
					Times(1)
			}

			_, err := api.Restore(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}
		})
	}
}
