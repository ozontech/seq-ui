package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/types"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

func TestStart(t *testing.T) {
	type mockArgs struct {
		resp types.StartExportResponse
		err  error
	}

	tests := []struct {
		name string

		req      *massexport_v1.StartRequest
		want     *massexport_v1.StartResponse
		wantCode codes.Code

		mockArgs *mockArgs
	}{
		{
			name: "ok",
			req: &massexport_v1.StartRequest{
				Query:  query,
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Window: testWindow,
				Name:   testName,
			},
			want: &massexport_v1.StartResponse{
				SessionId: sessionID,
			},
			mockArgs: &mockArgs{
				resp: types.StartExportResponse{
					SessionID: sessionID,
				},
			},
		},
		{
			name: "err_empty_name",
			req: &massexport_v1.StartRequest{
				Query:  query,
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Window: testWindow,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_invalid_window",
			req: &massexport_v1.StartRequest{
				Query:  query,
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Window: "invalid",
				Name:   testName,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_window_larger_than_interval",
			req: &massexport_v1.StartRequest{
				Query:  query,
				From:   timestamppb.New(from),
				To:     timestamppb.New(from.Add(5 * time.Second)),
				Window: testWindow,
				Name:   testName,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "err_svc",
			req: &massexport_v1.StartRequest{
				Query:  query,
				From:   timestamppb.New(from),
				To:     timestamppb.New(to),
				Window: testWindow,
				Name:   testName,
			},
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
					StartExport(gomock.Any(), gomock.Any()).
					Return(tt.mockArgs.resp, tt.mockArgs.err).
					Times(1)
			}

			got, err := api.Start(context.Background(), tt.req)

			require.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode != codes.OK {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
