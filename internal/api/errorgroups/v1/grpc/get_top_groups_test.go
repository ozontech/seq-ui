package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestGetTopGroups(t *testing.T) {
	var (
		env      = "test-env"
		source   = "test-source"
		duration = 2 * time.Minute
		someErr  = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetTopErrorGroupsRequest

		groups []types.TopErrorGroup
		total  uint64
		err    error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetTopGroupsRequest
		want    *errorgroups_v1.GetTopGroupsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetTopGroupsRequest{
				Env:       &env,
				Source:    &source,
				Duration:  durationpb.New(duration),
				Limit:     2,
				Offset:    0,
				WithTotal: true,
			},
			want: &errorgroups_v1.GetTopGroupsResponse{
				Total: 10,
				Groups: []*errorgroups_v1.GetTopGroupsResponse_Group{
					{
						Hash:      123,
						Message:   "some error 1",
						Source:    source,
						SeenTotal: 10,
					},
					{
						Hash:      456,
						Message:   "some error 2",
						Source:    source,
						SeenTotal: 5,
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Env:       &env,
					Source:    &source,
					Duration:  &duration,
					Limit:     2,
					Offset:    0,
					WithTotal: true,
				},

				groups: []types.TopErrorGroup{
					{
						Hash:    123,
						Message: "some error 1",
						Source:  source,
						Count:   10,
					},
					{
						Hash:    456,
						Message: "some error 2",
						Source:  source,
						Count:   5,
					},
				},
				total: 10,
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetTopGroupsRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{},

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
					GetTopErrorGroups(gomock.Any(), ma.req).
					Return(ma.groups, ma.total, ma.err).
					Times(1)
			}

			got, err := api.GetTopGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
