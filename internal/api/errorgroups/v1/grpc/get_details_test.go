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

func TestGetDetails(t *testing.T) {
	var (
		groupHash     = uint64(123)
		service       = "test-service"
		env           = "test-env"
		release       = "test-release"
		source        = "test-source"
		msg           = "some error"
		now           = time.Now()
		oneMinuteAgo  = now.Add(-1 * time.Minute)
		twoMinutesAgo = now.Add(-2 * time.Minute)
		someErr       = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorGroupDetailsRequest

		details types.ErrorGroupDetails
		err     error
	}

	tests := []struct {
		name string

		req     *errorgroups_v1.GetDetailsRequest
		want    *errorgroups_v1.GetDetailsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: &errorgroups_v1.GetDetailsRequest{
				GroupHash: groupHash,
				Service:   &service,
				Env:       &env,
				Release:   &release,
				Source:    &source,
			},
			want: &errorgroups_v1.GetDetailsResponse{
				GroupHash:   groupHash,
				Message:     msg,
				SeenTotal:   10,
				FirstSeenAt: timestamppb.New(twoMinutesAgo),
				LastSeenAt:  timestamppb.New(oneMinuteAgo),
				Source:      source,
				LogTags: map[string]string{
					"tag1": "val1",
					"tag2": "val2",
				},
				Distributions: &errorgroups_v1.GetDetailsResponse_Distributions{
					ByEnv: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "env1", Percent: 70},
						{Value: "env2", Percent: 30},
					},
					BySource: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "source1", Percent: 50},
						{Value: "source2", Percent: 50},
					},
					ByService: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "service1", Percent: 100},
						{Value: "service2", Percent: 0},
					},
					ByRelease: []*errorgroups_v1.GetDetailsResponse_Distribution{
						{Value: "release1", Percent: 60},
						{Value: "release2", Percent: 40},
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					Service:   &service,
					GroupHash: groupHash,
					Env:       &env,
					Release:   &release,
					Source:    &source,
				},

				details: types.ErrorGroupDetails{
					Hash:        groupHash,
					Message:     msg,
					SeenTotal:   10,
					FirstSeenAt: twoMinutesAgo,
					LastSeenAt:  oneMinuteAgo,
					Source:      source,
					LogTags: map[string]string{
						"tag1": "val1",
						"tag2": "val2",
					},
					Distributions: types.ErrorGroupDistributions{
						ByEnv: []types.ErrorGroupDistribution{
							{Value: "env1", Percent: 70},
							{Value: "env2", Percent: 30},
						},
						BySource: []types.ErrorGroupDistribution{
							{Value: "source1", Percent: 50},
							{Value: "source2", Percent: 50},
						},
						ByService: []types.ErrorGroupDistribution{
							{Value: "service1", Percent: 100},
							{Value: "service2", Percent: 0},
						},
						ByRelease: []types.ErrorGroupDistribution{
							{Value: "release1", Percent: 60},
							{Value: "release2", Percent: 40},
						},
					},
				},
			},
		},
		{
			name: "err_svc",

			req:     &errorgroups_v1.GetDetailsRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{},

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
					GetDetails(gomock.Any(), ma.req).
					Return(ma.details, ma.err).
					Times(1)
			}

			got, err := api.GetDetails(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
