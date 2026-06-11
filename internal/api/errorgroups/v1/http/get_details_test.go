package http

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	svc_mock "github.com/ozontech/seq-ui/internal/pkg/service/errorgroups/mock"
)

func TestServeGetDetails(t *testing.T) {
	var (
		service       = "test-service"
		groupHash     = uint64(123)
		msg           = "some error"
		env           = "test-env"
		source        = "test-source"
		release       = "test-release"
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

		req     getDetailsRequest
		want    getDetailsResponse
		wantErr bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: getDetailsRequest{
				GroupHash: fmt.Sprintf("%d", groupHash),
				Env:       &env,
				Source:    &source,
				Service:   &service,
				Release:   &release,
			},
			want: getDetailsResponse{
				GroupHash:   fmt.Sprintf("%d", groupHash),
				Message:     msg,
				Source:      source,
				SeenTotal:   10,
				FirstSeenAt: twoMinutesAgo,
				LastSeenAt:  oneMinuteAgo,
				LogTags: map[string]string{
					"tag1": "val1",
					"tag2": "val2",
				},
				Distributions: distributions{
					ByEnv: []distribution{
						{Value: "env1", Percent: 70},
						{Value: "env2", Percent: 30},
					},
					BySource: []distribution{
						{Value: "source1", Percent: 50},
						{Value: "source2", Percent: 50},
					},
					ByService: []distribution{
						{Value: "service1", Percent: 100},
						{Value: "service2", Percent: 0},
					},
					ByRelease: []distribution{
						{Value: "release1", Percent: 60},
						{Value: "release2", Percent: 40},
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: groupHash,
					Service:   &service,
					Env:       &env,
					Source:    &source,
					Release:   &release,
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

			req: getDetailsRequest{
				GroupHash: fmt.Sprintf("%d", groupHash),
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: groupHash,
				},

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

			httputil.DoTestHTTPEx(t, httputil.TestDataHTTPEx[getDetailsRequest, getDetailsResponse]{
				Method: http.MethodPost,
				Target: "/errorgroups/v1/details",
				Req:    tt.req,

				Handler: api.serveGetDetails,

				Want:    tt.want,
				WantErr: tt.wantErr,
			})
		})
	}
}
