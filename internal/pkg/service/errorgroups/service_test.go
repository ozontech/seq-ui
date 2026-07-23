package errorgroups

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	config "github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
)

func TestValidateTimeRange(t *testing.T) {
	var (
		dur  = time.Second
		to   = time.Now()
		from = to.Add(-2 * time.Second)
	)

	tests := []struct {
		name string

		tr      *types.TimeRange
		wantErr string
	}{
		{
			name: "ok_nil",
			tr:   nil,
		},
		{
			name: "ok_duration",
			tr: &types.TimeRange{
				Duration: dur,
			},
		},
		{
			name: "ok_from_to",
			tr: &types.TimeRange{
				From: from,
				To:   to,
			},
		},
		{
			name: "err_both",
			tr: &types.TimeRange{
				Duration: dur,
				From:     from,
				To:       to,
			},
			wantErr: "only one of 'duration' or 'from'/'to' must be specified",
		},
		{
			name:    "err_empty",
			tr:      &types.TimeRange{},
			wantErr: "at least one of 'duration' or 'from'/'to' must be specified",
		},
		{
			name: "err_only_from",
			tr: &types.TimeRange{
				From: from,
			},
			wantErr: "both 'from'/'to' must be specified",
		},
		{
			name: "err_only_to",
			tr: &types.TimeRange{
				To: to,
			},
			wantErr: "both 'from'/'to' must be specified",
		},
		{
			name: "err_from_to_equal",
			tr: &types.TimeRange{
				From: from,
				To:   from,
			},
			wantErr: "'from' should be before 'to'",
		},
		{
			name: "err_from_before_to",
			tr: &types.TimeRange{
				From: to,
				To:   from,
			},
			wantErr: "'from' should be before 'to'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateTimeRange(tt.tr)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGetErrorGroups(t *testing.T) {
	var (
		service = "test-svc"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorGroupsRequest

		groupsCount int
		errGroups   error

		total    uint64
		errTotal error
	}

	tests := []struct {
		name string

		req             types.GetErrorGroupsRequest
		wantGroupsCount int
		wantTotal       uint64
		wantErr         bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_limit",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Limit:   5,
			},
			wantGroupsCount: 5,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Limit:   5,
				},

				groupsCount: 5,
			},
		},
		{
			name: "ok_no_limit",

			req: types.GetErrorGroupsRequest{
				Service: service,
			},
			wantGroupsCount: 10,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Limit:   defaultLimit,
				},

				groupsCount: 10,
			},
		},
		{
			name: "ok_with_total",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Limit:     10,
				WithTotal: true,
			},
			wantGroupsCount: 10,
			wantTotal:       100,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				total:       100,
			},
		},
		{
			name: "ok_time_range",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Limit:   10,
				TimeRange: &types.TimeRange{
					Duration: time.Minute,
				},
			},
			wantGroupsCount: 10,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Limit:   10,
					TimeRange: &types.TimeRange{
						Duration: time.Minute,
					},
				},

				groupsCount: 10,
				total:       0,
			},
		},
		{
			name: "err_no_service",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,
		},
		{
			name: "err_timerange",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				TimeRange: &types.TimeRange{},
			},
			wantErr: true,
		},
		{
			name: "err_repo_groups",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Limit:     10,
					WithTotal: true,
				},

				errGroups: someErr,
				total:     100,
			},
		},
		{
			name: "err_repo_total",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				errTotal:    someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetErrorGroups(gomock.Any(), ma.req).
					Return(
						slices.Repeat([]types.ErrorGroup{{}}, ma.groupsCount),
						ma.errGroups,
					).
					Times(1)

				if tt.req.WithTotal {
					mockedRepo.EXPECT().
						GetErrorGroupsTotal(gomock.Any(), ma.req).
						Return(ma.total, ma.errTotal).
						Times(1)
				}
			}

			gotGroups, gotTotal, err := svc.GetErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantGroupsCount, len(gotGroups))
			require.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}

func TestGetNewErrorGroups(t *testing.T) {
	var (
		service  = "test-svc"
		release  = "test-release"
		duration = 24 * time.Hour
		someErr  = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorGroupsRequest

		groupsCount int
		errGroups   error

		total    uint64
		errTotal error
	}

	tests := []struct {
		name string

		req             types.GetErrorGroupsRequest
		wantGroupsCount int
		wantTotal       uint64
		wantErr         bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_empty_req",

			req:             types.GetErrorGroupsRequest{},
			wantGroupsCount: 0,
			wantTotal:       0,
		},
		{
			name: "ok_limit",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Release: &release,
				Limit:   5,
			},
			wantGroupsCount: 5,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Release: &release,
					Limit:   5,
				},

				groupsCount: 5,
			},
		},
		{
			name: "ok_no_limit",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Release: &release,
			},
			wantGroupsCount: 10,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Release: &release,
					Limit:   defaultLimit,
				},

				groupsCount: 10,
			},
		},
		{
			name: "ok_with_total",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Release:   &release,
				Limit:     10,
				WithTotal: true,
			},
			wantGroupsCount: 10,
			wantTotal:       100,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Release:   &release,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				total:       100,
			},
		},
		{
			name: "ok_timerange",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Release: &release,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit: 10,
			},
			wantGroupsCount: 10,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service: service,
					Release: &release,
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
					Limit: 10,
				},

				groupsCount: 10,
			},
		},
		{
			name: "err_no_service",

			req: types.GetErrorGroupsRequest{
				Release: &release,
			},
			wantErr: true,
		},
		{
			name: "err_timerange",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Release:   &release,
				TimeRange: &types.TimeRange{},
			},
			wantErr: true,
		},
		{
			name: "err_repo_groups",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Release:   &release,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Release:   &release,
					Limit:     10,
					WithTotal: true,
				},

				errGroups: someErr,
				total:     100,
			},
		},
		{
			name: "err_repo_total",

			req: types.GetErrorGroupsRequest{
				Service:   service,
				Release:   &release,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupsRequest{
					Service:   service,
					Release:   &release,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				errTotal:    someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetNewErrorGroups(gomock.Any(), ma.req).
					Return(
						slices.Repeat([]types.ErrorGroup{{}}, ma.groupsCount),
						ma.errGroups,
					).
					Times(1)

				if tt.req.WithTotal {
					mockedRepo.EXPECT().
						GetNewErrorGroupsTotal(gomock.Any(), ma.req).
						Return(ma.total, ma.errTotal).
						Times(1)
				}
			}

			gotGroups, gotTotal, err := svc.GetNewErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantGroupsCount, len(gotGroups))
			require.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}

func TestGetTopErrorGroups(t *testing.T) {
	var (
		duration = time.Hour
		someErr  = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetTopErrorGroupsRequest

		groupsCount int
		errGroups   error

		total    uint64
		errTotal error
	}

	tests := []struct {
		name string

		req             types.GetTopErrorGroupsRequest
		wantGroupsCount int
		wantTotal       uint64
		wantErr         bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_limit",

			req: types.GetTopErrorGroupsRequest{
				Limit: 5,
			},
			wantGroupsCount: 5,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit: 5,
				},

				groupsCount: 5,
			},
		},
		{
			name: "ok_no_limit",

			req:             types.GetTopErrorGroupsRequest{},
			wantGroupsCount: 10,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit: defaultLimit,
				},

				groupsCount: 10,
			},
		},
		{
			name: "ok_with_total",

			req: types.GetTopErrorGroupsRequest{
				Limit:     10,
				WithTotal: true,
			},
			wantGroupsCount: 10,
			wantTotal:       100,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				total:       100,
			},
		},
		{
			name: "ok_timerange",

			req: types.GetTopErrorGroupsRequest{
				Limit: 10,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},
			wantGroupsCount: 10,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit: 10,
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
				},

				groupsCount: 10,
			},
		},
		{
			name: "err_timerange",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{},
			},
			wantErr: true,
		},
		{
			name: "err_repo_groups",

			req: types.GetTopErrorGroupsRequest{
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit:     10,
					WithTotal: true,
				},

				errGroups: someErr,
				total:     100,
			},
		},
		{
			name: "err_repo_total",

			req: types.GetTopErrorGroupsRequest{
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetTopErrorGroupsRequest{
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				errTotal:    someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetTopErrorGroups(gomock.Any(), ma.req).
					Return(
						slices.Repeat([]types.TopErrorGroup{{}}, ma.groupsCount),
						ma.errGroups,
					).
					Times(1)

				if tt.req.WithTotal {
					mockedRepo.EXPECT().
						GetTopErrorGroupsTotal(gomock.Any(), ma.req).
						Return(ma.total, ma.errTotal).
						Times(1)
				}
			}

			gotGroups, gotTotal, err := svc.GetTopErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantGroupsCount, len(gotGroups))
			require.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}

func TestGetHist(t *testing.T) {
	var (
		duration = time.Hour
		someErr  = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorHistRequest

		bucketsCount int
		err          error
	}

	tests := []struct {
		name string

		req              types.GetErrorHistRequest
		wantBucketsCount int
		wantErr          bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req:              types.GetErrorHistRequest{},
			wantBucketsCount: 50,

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{},

				bucketsCount: 50,
			},
		},
		{
			name: "ok_timerange",

			req: types.GetErrorHistRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},
			wantBucketsCount: 50,

			mockArgs: &mockArgs{
				req: types.GetErrorHistRequest{
					TimeRange: &types.TimeRange{
						Duration: duration,
					},
				},

				bucketsCount: 50,
			},
		},
		{
			name: "err_timerange",

			req: types.GetErrorHistRequest{
				TimeRange: &types.TimeRange{},
			},
			wantErr: true,
		},
		{
			name: "err_repo",

			req:     types.GetErrorHistRequest{},
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
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetErrorHist(gomock.Any(), ma.req).
					Return(
						types.ErrorHist{
							Buckets: slices.Repeat([]types.ErrorHistBucket{{}}, ma.bucketsCount),
						},
						ma.err,
					).
					Times(1)
			}

			got, err := svc.GetHist(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantBucketsCount, len(got.Buckets))
		})
	}
}

func TestGetDetails(t *testing.T) {
	var (
		hash    = uint64(123)
		env     = "test-env"
		source  = "test-source"
		service = "test-svc"
		release = "test-release"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetErrorGroupDetailsRequest

		details    types.ErrorGroupDetails
		errDetails error

		counts    types.ErrorGroupCounts
		errCounts error
	}

	tests := []struct {
		name string

		req     types.GetErrorGroupDetailsRequest
		want    types.ErrorGroupDetails
		wantErr bool

		logTagsMapping config.LogTagsMapping

		mockArgs *mockArgs
	}{
		{
			name: "ok_full",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Source:    &source,
				Service:   &service,
				Release:   &release,
			},
			want: types.ErrorGroupDetails{
				SeenTotal: 10,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Source:    &source,
					Service:   &service,
					Release:   &release,
				},

				details: types.ErrorGroupDetails{
					SeenTotal: 10,
				},
			},
		},
		{
			name: "ok_not_full",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Service:   &service,
			},
			want: types.ErrorGroupDetails{
				SeenTotal: 10,
				Distributions: types.ErrorGroupDistributions{
					BySource: []types.ErrorGroupDistribution{
						{Value: "source2", Percent: 60},
						{Value: "source1", Percent: 40},
					},
					ByRelease: []types.ErrorGroupDistribution{
						{Value: "release1", Percent: 70},
						{Value: "release2", Percent: 30},
					},
				},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Service:   &service,
				},

				details: types.ErrorGroupDetails{
					SeenTotal: 10,
				},
				counts: types.ErrorGroupCounts{
					ByEnv:     types.ErrorGroupCount{env: 10},
					ByService: types.ErrorGroupCount{service: 10},
					BySource: types.ErrorGroupCount{
						"source1": 4,
						"source2": 6,
					},
					ByRelease: types.ErrorGroupCount{
						"release1": 7,
						"release2": 3,
					},
				},
			},
		},
		{
			name: "ok_not_full_seen_total_0",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Service:   &service,
			},
			want: types.ErrorGroupDetails{
				SeenTotal: 0,
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Service:   &service,
				},

				details: types.ErrorGroupDetails{
					SeenTotal: 0,
				},
			},
		},
		{
			name: "ok_log_tags_mapping",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
			},
			want: types.ErrorGroupDetails{
				LogTags: map[string]string{
					"hash":   "123",
					"source": source,
				},
			},

			logTagsMapping: config.LogTagsMapping{
				Env:     []string{"env"},
				Service: []string{"service"},
				Release: []string{"release", "app_version"},
			},

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
				},

				details: types.ErrorGroupDetails{
					LogTags: map[string]string{
						"hash":        "123",
						"env":         env,
						"source":      source,
						"service":     service,
						"app_version": release,
					},
				},
			},
		},
		{
			name: "err_full_repo",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Source:    &source,
				Service:   &service,
				Release:   &release,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Source:    &source,
					Service:   &service,
					Release:   &release,
				},

				errDetails: someErr,
			},
		},
		{
			name: "err_not_full_repo_details",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Service:   &service,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Service:   &service,
				},

				errDetails: someErr,
			},
		},
		{
			name: "err_not_full_repo_counts",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: hash,
				Env:       &env,
				Service:   &service,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetErrorGroupDetailsRequest{
					GroupHash: hash,
					Env:       &env,
					Service:   &service,
				},

				errCounts: someErr,
			},
		},
		{
			name: "err_no_hash",

			req:     types.GetErrorGroupDetailsRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, tt.logTagsMapping)

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetErrorDetails(gomock.Any(), ma.req).
					Return(ma.details, ma.errDetails).
					Times(1)

				if !tt.req.IsFullyFilled() {
					mockedRepo.EXPECT().
						GetErrorCounts(gomock.Any(), ma.req).
						Return(ma.counts, ma.errCounts).
						Times(1)
				}
			}

			got, err := svc.GetDetails(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetServices(t *testing.T) {
	var someErr = errors.New("some err")

	type mockArgs struct {
		req types.GetServicesRequest

		services []string
		err      error
	}

	tests := []struct {
		name string

		req          types.GetServicesRequest
		wantServices []string
		wantErr      bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req:          types.GetServicesRequest{},
			wantServices: []string{"service1", "service2"},

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{},

				services: []string{"service1", "service2"},
			},
		},
		{
			name: "err_repo",

			req:     types.GetServicesRequest{},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetServicesRequest{},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetServices(gomock.Any(), ma.req).
					Return(ma.services, ma.err).
					Times(1)
			}

			got, err := svc.GetServices(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantServices, got)
		})
	}
}

func TestGetReleases(t *testing.T) {
	var (
		service = "test-svc"
		someErr = errors.New("some err")
	)

	type mockArgs struct {
		req types.GetReleasesRequest

		releases []string
		err      error
	}

	tests := []struct {
		name string

		req          types.GetReleasesRequest
		wantReleases []string
		wantErr      bool

		mockArgs *mockArgs
	}{
		{
			name: "ok",

			req: types.GetReleasesRequest{
				Service: service,
			},
			wantReleases: []string{"release1", "release2"},

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{
					Service: service,
				},

				releases: []string{"release1", "release2"},
			},
		},
		{
			name: "err_no_service",

			req:     types.GetReleasesRequest{},
			wantErr: true,
		},
		{
			name: "err_repo",

			req: types.GetReleasesRequest{
				Service: service,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.GetReleasesRequest{
					Service: service,
				},

				err: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					GetReleases(gomock.Any(), ma.req).
					Return(ma.releases, ma.err).
					Times(1)
			}

			got, err := svc.GetReleases(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantReleases, got)
		})
	}
}

func TestDiffByReleases(t *testing.T) {
	var (
		service  = "test-svc"
		releases = []string{"release1", "release2"}
		someErr  = errors.New("some err")
	)

	type mockArgs struct {
		req types.DiffByReleasesRequest

		groupsCount int
		errGroups   error

		total    uint64
		errTotal error
	}

	tests := []struct {
		name string

		req             types.DiffByReleasesRequest
		wantGroupsCount int
		wantTotal       uint64
		wantErr         bool

		mockArgs *mockArgs
	}{
		{
			name: "ok_limit",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
				Limit:    5,
			},
			wantGroupsCount: 5,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:  service,
					Releases: releases,
					Limit:    5,
				},

				groupsCount: 5,
			},
		},
		{
			name: "ok_no_limit",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
			},
			wantGroupsCount: 10,
			wantTotal:       0,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:  service,
					Releases: releases,
					Limit:    defaultLimit,
				},

				groupsCount: 10,
			},
		},
		{
			name: "ok_with_total",

			req: types.DiffByReleasesRequest{
				Service:   service,
				Releases:  releases,
				Limit:     10,
				WithTotal: true,
			},
			wantGroupsCount: 10,
			wantTotal:       100,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:   service,
					Releases:  releases,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				total:       100,
			},
		},
		{
			name: "err_no_service",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,
		},
		{
			name: "err_not_enough_releases",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: []string{"release1"},
			},
			wantErr: true,
		},
		{
			name: "err_empty_release",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: []string{"release1", ""},
			},
			wantErr: true,
		},
		{
			name: "err_repo_groups",

			req: types.DiffByReleasesRequest{
				Service:   service,
				Releases:  releases,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:   service,
					Releases:  releases,
					Limit:     10,
					WithTotal: true,
				},

				errGroups: someErr,
				total:     100,
			},
		},
		{
			name: "err_repo_total",

			req: types.DiffByReleasesRequest{
				Service:   service,
				Releases:  releases,
				Limit:     10,
				WithTotal: true,
			},
			wantErr: true,

			mockArgs: &mockArgs{
				req: types.DiffByReleasesRequest{
					Service:   service,
					Releases:  releases,
					Limit:     10,
					WithTotal: true,
				},

				groupsCount: 10,
				errTotal:    someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedRepo := mock.NewMockRepository(ctrl)

			svc := New(mockedRepo, config.LogTagsMapping{})

			if ma := tt.mockArgs; ma != nil {
				mockedRepo.EXPECT().
					DiffByReleases(gomock.Any(), ma.req).
					Return(
						slices.Repeat([]types.DiffGroup{{}}, ma.groupsCount),
						ma.errGroups,
					).
					Times(1)

				if tt.req.WithTotal {
					mockedRepo.EXPECT().
						DiffByReleasesTotal(gomock.Any(), ma.req).
						Return(ma.total, ma.errTotal).
						Times(1)
				}
			}

			gotGroups, gotTotal, err := svc.DiffByReleases(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantGroupsCount, len(gotGroups))
			require.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}
