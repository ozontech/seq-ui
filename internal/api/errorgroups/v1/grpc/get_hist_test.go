package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetHist(t *testing.T) {
	var (
		service       = "test-service"
		groupHash     = uint64(123)
		env           = "test-env"
		source        = "test-source"
		release       = "test-release"
		duration      = 2 * time.Minute
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorHistRequest

		hist types.ErrorHist
		err  error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetHistRequest
		want    *errorgroups_v1.GetHistResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetHistRequest{
				GroupHash: &groupHash,
				Service:   &service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				Duration:  durationpb.New(duration),
			},
			want: &errorgroups_v1.GetHistResponse{
				Buckets: []*errorgroups_v1.Bucket{
					{Time: timestamppb.New(oneMinuteAgo), Count: 10},
					{Time: timestamppb.New(twoMinutesAgo), Count: 20},
				},
				Interval: 123,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					GroupHash: &groupHash,
					Service:   &service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
					Duration:  &duration,
				},

				hist: types.ErrorHist{
					Buckets: []types.ErrorHistBucket{
						{Time: oneMinuteAgo, Count: 10},
						{Time: twoMinutesAgo, Count: 20},
					},
					Interval: 123,
				},
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetHistRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{},

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
					GetHist(gomock.Any(), ma.req).
					Return(ma.hist, ma.err).
					Times(1)
			}

			got, err := api.GetHist(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
