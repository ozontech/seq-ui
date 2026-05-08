package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
	errorgroups_v1 "github.com/ozontech/seq-ui/pkg/errorgroups/v1"
)

func TestDiffByReleases(t *testing.T) {
	var (
		service       = "test-service"
		env           = "test-env"
		source        = "test-source"
		releases      = []string{"release1", "release2"}
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.DiffByReleasesRequest

		groups []types.DiffGroup
		total  uint64
		err    error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.DiffByReleasesRequest
		want    *errorgroups_v1.DiffByReleasesResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.DiffByReleasesRequest{
				Service:   service,
				Releases:  releases,
				Env:       &env,
				Source:    &source,
				Limit:     2,
				Offset:    0,
				Order:     errorgroups_v1.Order_ORDER_LATEST,
				WithTotal: true,
			},
			want: &errorgroups_v1.DiffByReleasesResponse{
				Total: 10,
				Groups: []*errorgroups_v1.DiffByReleasesResponse_Group{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
						ReleaseInfos: map[string]*errorgroups_v1.DiffByReleasesResponse_ReleaseInfo{
							"release1": {SeenTotal: 10},
							"release2": {SeenTotal: 20},
						},
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						FirstSeenAt: timestamppb.New(twoMinutesAgo),
						LastSeenAt:  timestamppb.New(oneMinuteAgo),
						ReleaseInfos: map[string]*errorgroups_v1.DiffByReleasesResponse_ReleaseInfo{
							"release1": {SeenTotal: 30},
							"release2": {SeenTotal: 0},
						},
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:   service,
					Releases:  releases,
					Env:       &env,
					Source:    &source,
					Limit:     2,
					Offset:    0,
					Order:     types.OrderLatest,
					WithTotal: true,
				},

				groups: []types.DiffGroup{
					{
						Hash:        123,
						Message:     "some error 1",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]types.DiffReleaseInfo{
							"release1": {SeenTotal: 10},
							"release2": {SeenTotal: 20},
						},
					},
					{
						Hash:        456,
						Message:     "some error 2",
						Source:      source,
						FirstSeenAt: twoMinutesAgo,
						LastSeenAt:  oneMinuteAgo,
						ReleaseInfos: map[string]types.DiffReleaseInfo{
							"release1": {SeenTotal: 30},
							"release2": {SeenTotal: 0},
						},
					},
				},
				total: 10,
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.DiffByReleasesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedSvc := svc_mock.NewMockService(ctrl)

			api := New(mockedSvc)

			if ma := tt.mockArgs; ma != nil {
				mockedSvc.EXPECT().
					DiffByReleases(gomock.Any(), ma.req).
					Return(ma.groups, ma.total, ma.err).
					Times(1)
			}

			got, err := api.DiffByReleases(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
