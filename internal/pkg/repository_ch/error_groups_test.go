package repositorych

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ozontech/seq-ui/internal/app/types"
	mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func fakeNow(now time.Time) func() time.Time {
	return func() time.Time {
		return now
	}
}

func TestGetNewErrorGroups(t *testing.T) {
	var (
		service = "test-svc"
		release = "test-release"
		env     = "test-env"
		source  = "test-source"

		fakeNow  = fakeNow(time.Now())
		duration = time.Hour * 24
		timeDiff = fakeNow().Add(-duration.Abs())

		someErr = errors.New("some err")
	)

	type mockRows struct {
		count   int
		scanErr error
	}

	type mockConn struct {
		query string
		args  []any

		rows *mockRows
		err  error
	}

	tests := []struct {
		name string

		req        types.GetErrorGroupsRequest
		wantGroups int
		wantErr    bool

		isSharded   bool
		queryFilter map[string]string
		mockConn    *mockConn
	}{
		{
			name: "ok_by_releases",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Release:  &release,
				Duration: &duration,
				Limit:    20,
				Offset:   5,
				Order:    types.OrderFrequent,
			},
			wantGroups: 2,

			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
					" FROM error_groups"+
					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
					" GROUP BY _group_hash, source"+
					" ORDER BY seen_total DESC",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING any(release) = ? AND count() = ?"+
						" ORDER BY countMerge(seen_total) DESC"+
						" LIMIT 20 OFFSET 5",
				),
				args: []any{service, release, 1},

				rows: &mockRows{
					count: 2,
				},
			},
		},
		{
			name: "ok_by_duration",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Duration: &duration,
				Limit:    10,
				Order:    types.OrderLatest,
			},
			wantGroups: 2,

			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
					" FROM error_groups"+
					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
					" GROUP BY _group_hash, source"+
					" ORDER BY last_seen_at DESC",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING minMerge(first_seen_at) >= ?"+
						" ORDER BY maxMerge(last_seen_at) DESC"+
						" LIMIT 10 OFFSET 0",
				),
				args: []any{service, timeDiff},

				rows: &mockRows{
					count: 2,
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Env:      &env,
				Source:   &source,
				Duration: &duration,
				Limit:    10,
				Offset:   20,
				Order:    types.OrderOldest,
			},
			wantGroups: 2,

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},
			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
					" FROM error_groups"+
					" WHERE _group_hash IN (%s) AND service = 'test-svc' AND filter1 = 'value1' AND filter2 = 'value2' AND source = 'test-source' AND env = 'test-env'"+
					" GROUP BY _group_hash, source"+
					" ORDER BY first_seen_at",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING minMerge(first_seen_at) >= ?"+
						" ORDER BY minMerge(first_seen_at)"+
						" LIMIT 10 OFFSET 20",
				),
				args: []any{service, "value1", "value2", env, source, timeDiff},

				rows: &mockRows{
					count: 2,
				},
			},
		},
		{
			name: "ok_sharded",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Duration: &duration,
				Limit:    10,
				Offset:   20,
				Order:    types.OrderFrequent,
			},
			wantGroups: 2,

			isSharded: true,
			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
					" FROM error_groups"+
					" WHERE _group_hash GLOBAL IN (%s) AND service = 'test-svc'"+
					" GROUP BY _group_hash, source"+
					" ORDER BY seen_total DESC",

					"SELECT DISTINCT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING minMerge(first_seen_at) >= ?"+
						" ORDER BY countMerge(seen_total) DESC"+
						" LIMIT 10 OFFSET 20",
				),
				args: []any{service, timeDiff},

				rows: &mockRows{
					count: 2,
				},
			},
		},
		{
			name: "ok_no_rows",

			req:        types.GetErrorGroupsRequest{},
			wantGroups: 0,

			mockConn: &mockConn{
				rows: &mockRows{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConn{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConn{
				rows: &mockRows{
					scanErr: someErr,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedConn := mock.NewMockConn(ctrl)

			repo := newRepo(mockedConn, tt.isSharded, tt.queryFilter, fakeNow)

			if tt.mockConn != nil {
				mockedRows := mock.NewMockRows(ctrl)
				if rows := tt.mockConn.rows; rows != nil {
					times := rows.count
					if rows.scanErr != nil {
						times = 1
					}

					mockedRows.EXPECT().Next().Return(true).Times(times)
					mockedRows.EXPECT().Scan(gomock.Any()).Return(rows.scanErr).Times(times)
					if rows.scanErr == nil {
						mockedRows.EXPECT().Next().Return(false).Times(1)
					}
				}

				if tt.mockConn.query == "" {
					mockedConn.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockedRows, tt.mockConn.err).Times(1)
				} else {
					mockedConn.EXPECT().Query(gomock.Any(), tt.mockConn.query, tt.mockConn.args...).Return(mockedRows, tt.mockConn.err).Times(1)
				}
			}

			got, err := repo.GetNewErrorGroups(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantGroups, len(got))
		})
	}
}

func TestGetNewErrorGroupsCount(t *testing.T) {
	var (
		service = "test-svc"
		release = "test-release"
		env     = "test-env"
		source  = "test-source"

		fakeNow  = fakeNow(time.Now())
		duration = time.Hour * 24
		timeDiff = fakeNow().Add(-duration.Abs())

		someErr = errors.New("some err")
	)

	type mockConn struct {
		query string
		args  []any

		scanErr error
	}

	tests := []struct {
		name string

		req     types.GetErrorGroupsRequest
		wantErr bool

		queryFilter map[string]string
		mockConn    *mockConn
	}{
		{
			name: "ok_by_releases",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Release:  &release,
				Duration: &duration,
			},

			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING any(release) = ? AND count() = ?",
				),
				args: []any{service, release, 1},
			},
		},
		{
			name: "ok_by_duration",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Duration: &duration,
			},

			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING minMerge(first_seen_at) >= ?",
				),
				args: []any{service, timeDiff},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupsRequest{
				Service:  service,
				Duration: &duration,
				Env:      &env,
				Source:   &source,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},
			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
						" GROUP BY _group_hash, source"+
						" HAVING minMerge(first_seen_at) >= ?",
				),
				args: []any{service, "value1", "value2", env, source, timeDiff},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetErrorGroupsRequest{},

			mockConn: &mockConn{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConn{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedConn := mock.NewMockConn(ctrl)

			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			if tt.mockConn != nil {
				mockedRow := mock.NewMockRow(ctrl)
				mockedRow.EXPECT().Scan(gomock.Any()).Return(tt.mockConn.scanErr)

				if tt.mockConn.query == "" {
					mockedConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockedRow).Times(1)
				} else {
					mockedConn.EXPECT().QueryRow(gomock.Any(), tt.mockConn.query, tt.mockConn.args...).Return(mockedRow).Times(1)
				}
			}

			got, err := repo.GetNewErrorGroupsCount(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				require.Equal(t, uint64(0), got)
			}
		})
	}
}

func TestDiffByReleases(t *testing.T) {
	var (
		service  = "test-svc"
		releases = []string{"test-release1", "test-release2"}
		env      = "test-env"
		source   = "test-source"

		someErr = errors.New("some err")
	)

	type mockRows struct {
		scanFns []func(...any) error
		scanErr bool
	}

	type mockConn struct {
		query string
		args  []any

		rows *mockRows
		err  error
	}

	tests := []struct {
		name string

		req        types.DiffByReleasesRequest
		wantGroups []types.DiffGroup
		wantErr    bool

		queryFilter              map[string]string
		mockConnGroups, mockConn *mockConn
	}{
		{
			name: "ok",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
				Limit:    20,
				Order:    types.OrderFrequent,
			},
			wantGroups: []types.DiffGroup{
				{
					Hash: 123,
					ReleaseInfos: map[string]types.DiffReleaseInfo{
						releases[0]: {SeenTotal: 10},
						releases[1]: {SeenTotal: 20},
					},
				},
				{
					Hash: 456,
					ReleaseInfos: map[string]types.DiffReleaseInfo{
						releases[0]: {SeenTotal: 0},
						releases[1]: {SeenTotal: 1000},
					},
				},
			},

			mockConnGroups: &mockConn{
				query: "" +
					"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
					" FROM error_groups" +
					" WHERE release IN (?,?) AND service = ?" +
					" GROUP BY _group_hash, source" +
					" ORDER BY countMerge(seen_total) DESC" +
					" LIMIT 20 OFFSET 0",
				args: []any{releases[0], releases[1], service},

				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 456
							return nil
						},
					},
				},
			},
			mockConn: &mockConn{
				query: "" +
					"SELECT _group_hash, release, countMerge(seen_total) as seen_total" +
					" FROM error_groups" +
					" WHERE _group_hash IN (?,?) AND release IN (?,?) AND service = ?" +
					" GROUP BY _group_hash, release",
				args: []any{uint64(123), uint64(456), releases[0], releases[1], service},

				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							*args[1].(*string) = releases[0]
							*args[2].(*uint64) = 10
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 123
							*args[1].(*string) = releases[1]
							*args[2].(*uint64) = 20
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 456
							*args[1].(*string) = releases[1]
							*args[2].(*uint64) = 1000
							return nil
						},
					},
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
				Env:      &env,
				Source:   &source,
				Limit:    20,
				Offset:   5,
				Order:    types.OrderLatest,
			},
			wantGroups: []types.DiffGroup{
				{
					Hash: 123,
					ReleaseInfos: map[string]types.DiffReleaseInfo{
						releases[0]: {SeenTotal: 10},
						releases[1]: {SeenTotal: 20},
					},
				},
				{
					Hash: 456,
					ReleaseInfos: map[string]types.DiffReleaseInfo{
						releases[0]: {SeenTotal: 0},
						releases[1]: {SeenTotal: 1000},
					},
				},
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},
			mockConnGroups: &mockConn{
				query: "" +
					"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
					" FROM error_groups" +
					" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release IN (?,?) AND service = ? AND source = ?" +
					" GROUP BY _group_hash, source" +
					" ORDER BY last_seen_at DESC" +
					" LIMIT 20 OFFSET 5",
				args: []any{env, "value1", "value2", releases[0], releases[1], service, source},

				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 456
							return nil
						},
					},
				},
			},
			mockConn: &mockConn{
				query: "" +
					"SELECT _group_hash, release, countMerge(seen_total) as seen_total" +
					" FROM error_groups" +
					" WHERE _group_hash IN (?,?) AND env = ? AND filter1 = ? AND filter2 = ? AND release IN (?,?) AND service = ? AND source = ?" +
					" GROUP BY _group_hash, release",
				args: []any{uint64(123), uint64(456), env, "value1", "value2", releases[0], releases[1], service, source},

				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							*args[1].(*string) = releases[0]
							*args[2].(*uint64) = 10
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 123
							*args[1].(*string) = releases[1]
							*args[2].(*uint64) = 20
							return nil
						},
						func(args ...any) error {
							*args[0].(*uint64) = 456
							*args[1].(*string) = releases[1]
							*args[2].(*uint64) = 1000
							return nil
						},
					},
				},
			},
		},
		{
			name: "ok_no_rows",

			req:        types.DiffByReleasesRequest{},
			wantGroups: nil,

			mockConnGroups: &mockConn{
				rows: &mockRows{},
			},
		},
		{
			name: "err_query_groups",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConnGroups: &mockConn{
				err: someErr,
			},
		},
		{
			name: "err_scan_groups",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConnGroups: &mockConn{
				rows: &mockRows{
					scanErr: true,
					scanFns: []func(...any) error{
						func(args ...any) error { return someErr },
					},
				},
			},
		},
		{
			name: "err_query",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConnGroups: &mockConn{
				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							return nil
						},
					},
				},
			},
			mockConn: &mockConn{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConnGroups: &mockConn{
				rows: &mockRows{
					scanFns: []func(...any) error{
						func(args ...any) error {
							*args[0].(*uint64) = 123
							return nil
						},
					},
				},
			},
			mockConn: &mockConn{
				rows: &mockRows{
					scanErr: true,
					scanFns: []func(...any) error{
						func(args ...any) error { return someErr },
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedConn := mock.NewMockConn(ctrl)

			repo := newRepo(mockedConn, true, tt.queryFilter, time.Now)

			initMockConn := func(mc *mockConn) {
				if mc == nil {
					return
				}

				mockedRows := mock.NewMockRows(ctrl)
				if rows := mc.rows; rows != nil {
					for _, scanFn := range rows.scanFns {
						mockedRows.EXPECT().Next().Return(true)
						mockedRows.EXPECT().Scan(gomock.Any()).DoAndReturn(scanFn)
					}
					if !rows.scanErr {
						mockedRows.EXPECT().Next().Return(false)
					}
				}

				if mc.query == "" {
					mockedConn.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockedRows, mc.err).Times(1)
				} else {
					mockedConn.EXPECT().Query(gomock.Any(), mc.query, mc.args...).Return(mockedRows, mc.err).Times(1)
				}
			}

			initMockConn(tt.mockConnGroups)
			initMockConn(tt.mockConn)

			got, err := repo.DiffByReleases(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.wantGroups, got)
		})
	}
}

func TestDiffByReleasesTotal(t *testing.T) {
	var (
		service  = "test-svc"
		releases = []string{"test-release1", "test-release2"}
		env      = "test-env"
		source   = "test-source"

		someErr = errors.New("some err")
	)

	type mockConn struct {
		query string
		args  []any

		scanErr error
	}

	tests := []struct {
		name string

		req     types.DiffByReleasesRequest
		wantErr bool

		queryFilter map[string]string
		mockConn    *mockConn
	}{
		{
			name: "ok",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
			},

			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE release IN (?,?) AND service = ?"+
						" GROUP BY _group_hash",
				),
				args: []any{releases[0], releases[1], service},
			},
		},
		{
			name: "ok_full_filters",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
				Env:      &env,
				Source:   &source,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},
			mockConn: &mockConn{
				query: fmt.Sprintf(""+
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE release IN (?,?) AND service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
						" GROUP BY _group_hash",
				),
				args: []any{releases[0], releases[1], service, "value1", "value2", env, source},
			},
		},
		{
			name: "ok_no_rows",

			req: types.DiffByReleasesRequest{},

			mockConn: &mockConn{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConn: &mockConn{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockedConn := mock.NewMockConn(ctrl)

			repo := newRepo(mockedConn, true, tt.queryFilter, time.Now)

			if tt.mockConn != nil {
				mockedRow := mock.NewMockRow(ctrl)
				mockedRow.EXPECT().Scan(gomock.Any()).Return(tt.mockConn.scanErr)

				if tt.mockConn.query == "" {
					mockedConn.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockedRow).Times(1)
				} else {
					mockedConn.EXPECT().QueryRow(gomock.Any(), tt.mockConn.query, tt.mockConn.args...).Return(mockedRow).Times(1)
				}
			}

			got, err := repo.DiffByReleasesTotal(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				require.Equal(t, uint64(0), got)
			}
		})
	}
}
