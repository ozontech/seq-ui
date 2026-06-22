package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/app/types"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

func TestCheck(t *testing.T) {
	type mockArgs struct {
		resp types.ExportInfo
		err  error
	}

	tests := []struct {
		name string

		req      *massexport_v1.CheckRequest
		want     *massexport_v1.CheckResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req:  &massexport_v1.CheckRequest{SessionId: sessionID},
			want: convertExportInfo(types.ExportInfo{
				ID:     sessionID,
				UserID: userID,
				Status: types.ExportStatusFinish,
			}),
			mockArgs: &mockArgs{
				resp: types.ExportInfo{
					ID:     sessionID,
					UserID: userID,
					Status: types.ExportStatusFinish,
				},
			},
		},
		{
			name:     "err_svc",
			req:      &massexport_v1.CheckRequest{SessionId: sessionID},
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
					CheckExport(gomock.Any(), tt.req.GetSessionId()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.Check(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
