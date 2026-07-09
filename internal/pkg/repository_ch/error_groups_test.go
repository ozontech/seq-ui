package repositorych

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	sq "github.com/n-r-w/squirrel"
	"github.com/stretchr/testify/require"

	"github.com/ozontech/seq-ui/internal/app/types"
)

func TestGetHistData(t *testing.T) {
	const (
		_10min = 10 * time.Minute
		hour   = time.Hour
		day    = 24 * time.Hour
		week   = 7 * day
		month  = 31 * day
	)

	fakeNow := fakeNow(time.Now())

	tests := []struct {
		name string

		tr   *types.TimeRange
		want histData
	}{
		{
			name: "nil",

			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfMonth(start_date)",
				interval: uint64(month.Seconds()),
			},
		},
		{
			name: "empty",

			tr: &types.TimeRange{},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfMonth(start_date)",
				interval: uint64(month.Seconds()),
			},
		},
		{
			name: "duration_5_hour",

			tr: &types.TimeRange{
				Duration: 5 * time.Hour,
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "start_date",
				interval: uint64(_10min.Seconds()),
			},
		},
		{
			name: "duration_1_day",

			tr: &types.TimeRange{
				Duration: day,
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "toStartOfHour(start_date)",
				interval: uint64(hour.Seconds()),
			},
		},
		{
			name: "duration_1_month",

			tr: &types.TimeRange{
				Duration: month,
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "toStartOfDay(start_date)",
				interval: uint64(day.Seconds()),
			},
		},
		{
			name: "duration_7_month",

			tr: &types.TimeRange{
				Duration: 7 * month,
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfWeek(start_date)",
				interval: uint64(week.Seconds()),
			},
		},
		{
			name: "duration_1_year",

			tr: &types.TimeRange{
				Duration: 12 * month,
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfMonth(start_date)",
				interval: uint64(month.Seconds()),
			},
		},
		{
			name: "absolute_old_1_month",

			tr: &types.TimeRange{
				From: fakeNow().Add(-(3*month + 14*day)),
				To:   fakeNow().Add(-(2*month + 14*day)),
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "start_date",
				interval: uint64(day.Seconds()),
			},
		},
		{
			name: "absolute_old_6_month",

			tr: &types.TimeRange{
				From: fakeNow().Add(-12 * month),
				To:   fakeNow().Add(-6 * month),
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfWeek(start_date)",
				interval: uint64(week.Seconds()),
			},
		},
		{
			name: "absolute_old_1_year",

			tr: &types.TimeRange{
				From: fakeNow().Add(-18 * month),
				To:   fakeNow().Add(-6 * month),
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfMonth(start_date)",
				interval: uint64(month.Seconds()),
			},
		},
		{
			name: "absolute_new_5_hour",

			tr: &types.TimeRange{
				From: fakeNow().Add(-5 * time.Hour),
				To:   fakeNow(),
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "start_date",
				interval: uint64(_10min.Seconds()),
			},
		},
		{
			name: "absolute_new_1_day",

			tr: &types.TimeRange{
				From: fakeNow().Add(-day),
				To:   fakeNow(),
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "toStartOfHour(start_date)",
				interval: uint64(hour.Seconds()),
			},
		},
		{
			name: "absolute_new_1_month",

			tr: &types.TimeRange{
				From: fakeNow().Add(-month),
				To:   fakeNow(),
			},
			want: histData{
				table:    "agg_events_10min",
				column:   "toStartOfDay(start_date)",
				interval: uint64(day.Seconds()),
			},
		},
		{
			name: "absolute_new_2_month",

			tr: &types.TimeRange{
				From: fakeNow().Add(-2 * month),
				To:   fakeNow(),
			},
			want: histData{
				table:    "agg_events_1d",
				column:   "toStartOfWeek(start_date)",
				interval: uint64(week.Seconds()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := newRepo(nil, false, nil, fakeNow)

			got := r.getHistData(tt.tr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTimeRangeCond(t *testing.T) {
	var (
		fakeNow = fakeNow(time.Now())
		col     = "test-col"
	)

	tests := []struct {
		name string

		tr   *types.TimeRange
		want any
	}{
		{
			name: "nil",

			want: nil,
		},
		{
			name: "absolute",

			tr: &types.TimeRange{
				From: fakeNow().Add(-time.Hour),
				To:   fakeNow(),
			},
			want: sq.And{
				sq.GtOrEq{col: fakeNow().Add(-time.Hour)},
				sq.LtOrEq{col: fakeNow()},
			},
		},
		{
			name: "relative",

			tr: &types.TimeRange{
				Duration: time.Hour,
			},
			want: sq.GtOrEq{col: fakeNow().Add(-time.Hour)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := newRepo(nil, false, nil, fakeNow)

			got := r.timeRangeCond(col, tt.tr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetErrorGroups(t *testing.T) {
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

	tests := []struct {
		name string

		req             types.GetErrorGroupsRequest
		wantGroupsCount int
		wantErr         bool

		isSharded   bool
		queryFilter map[string]string

		mockConns []*mockConnRows
	}{
		{
			name: "ok_no_timerange",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Limit:   10,
				Offset:  20,
				Order:   types.OrderFrequent,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
							" FROM error_groups"+
							" WHERE service = ? AND _group_hash IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY seen_total DESC",

						"SELECT _group_hash"+
							" FROM error_groups"+
							" WHERE service = ?"+
							" GROUP BY _group_hash"+
							" ORDER BY countMerge(seen_total) DESC"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{
						service,
						service,
					},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_relative_frequent",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit:  10,
				Offset: 20,
				Order:  types.OrderFrequent,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE service = ? AND toStartOfHour(start_date) >= ?" +
						" GROUP BY _group_hash" +
						" ORDER BY count DESC" +
						" LIMIT 10 OFFSET 20",
					args: []any{service, timeDiff},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?) AND service = ?" +
						" GROUP BY _group_hash, source",
					args: []any{uint64(123), uint64(456), service},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_absolute_frequent",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
				Limit:  10,
				Offset: 20,
				Order:  types.OrderFrequent,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE service = ? AND (toStartOfHour(start_date) >= ? AND toStartOfHour(start_date) <= ?)" +
						" GROUP BY _group_hash" +
						" ORDER BY count DESC" +
						" LIMIT 10 OFFSET 20",
					args: []any{service, timeDiff, fakeNow()},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?) AND service = ?" +
						" GROUP BY _group_hash, source",
					args: []any{uint64(123), uint64(456), service},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_relative_not_frequent",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit:  10,
				Offset: 20,
				Order:  types.OrderLatest,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
							" FROM error_groups"+
							" WHERE service = ? AND _group_hash IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY last_seen_at DESC",

						"SELECT _group_hash"+
							" FROM error_groups"+
							" WHERE service = ?"+
							" GROUP BY _group_hash"+
							" HAVING maxMerge(last_seen_at) >= ?"+
							" ORDER BY maxMerge(last_seen_at) DESC"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{
						service,
						service, timeDiff,
					},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorInfo) = errorInfo{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorInfo) = errorInfo{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE _group_hash IN (?,?) AND service = ? AND toStartOfHour(start_date) >= ?" +
						" GROUP BY _group_hash",
					args: []any{uint64(123), uint64(456), service, timeDiff},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_absolute_not_frequent",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
				Limit:  10,
				Offset: 20,
				Order:  types.OrderLatest,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
							" FROM error_groups"+
							" WHERE service = ? AND _group_hash IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY last_seen_at DESC",

						"SELECT _group_hash"+
							" FROM error_groups"+
							" WHERE service = ?"+
							" GROUP BY _group_hash"+
							" HAVING (maxMerge(last_seen_at) >= ? AND maxMerge(last_seen_at) <= ?)"+
							" ORDER BY maxMerge(last_seen_at) DESC"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{
						service,
						service, timeDiff, fakeNow(),
					},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorInfo) = errorInfo{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorInfo) = errorInfo{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE _group_hash IN (?,?) AND service = ? AND (toStartOfHour(start_date) >= ? AND toStartOfHour(start_date) <= ?)" +
						" GROUP BY _group_hash",
					args: []any{uint64(123), uint64(456), service, timeDiff, fakeNow()},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_full_filters_sharded",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Env:     &env,
				Source:  &source,
				Release: &release,
				Limit:   10,
				Offset:  20,
				Order:   types.OrderOldest,
			},
			wantGroupsCount: 2,

			isSharded: true,
			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
							" FROM error_groups"+
							" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release = ? AND service = ? AND source = ? AND _group_hash GLOBAL IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY first_seen_at",

						"SELECT DISTINCT _group_hash"+
							" FROM error_groups"+
							" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release = ? AND service = ? AND source = ?"+
							" GROUP BY _group_hash"+
							" ORDER BY minMerge(first_seen_at)"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{
						env, "value1", "value2", release, service, source,
						env, "value1", "value2", release, service, source,
					},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_no_rows_no_duration",

			req:             types.GetErrorGroupsRequest{},
			wantGroupsCount: 0,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "ok_no_rows_timerange_frequent",

			req: types.GetErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Order: types.OrderFrequent,
			},
			wantGroupsCount: 0,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "ok_no_rows_timerange_no_frequent",

			req: types.GetErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Order: types.OrderLatest,
			},
			wantGroupsCount: 0,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					err: someErr,
				},
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						scanErr:      someErr,
						isScanStruct: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConns...)
			repo := newRepo(mockedConn, tt.isSharded, tt.queryFilter, fakeNow)

			got, err := repo.GetErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantGroupsCount, len(got))
		})
	}
}

func TestGetErrorGroupsTotal(t *testing.T) {
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

	tests := []struct {
		name string

		req     types.GetErrorGroupsRequest
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRow
	}{
		{
			name: "ok",

			req: types.GetErrorGroupsRequest{
				Service: service,
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups" +
					" WHERE service = ?",

				args: []any{service},
			},
		},
		{
			name: "ok_timerange_relative",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM agg_events_10min" +
					" WHERE service = ? AND toStartOfHour(start_date) >= ?",

				args: []any{service, timeDiff},
			},
		},
		{
			name: "ok_timerange_absolute",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM agg_events_10min" +
					" WHERE service = ? AND (toStartOfHour(start_date) >= ? AND toStartOfHour(start_date) <= ?)",

				args: []any{service, timeDiff, fakeNow()},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Env:     &env,
				Source:  &source,
				Release: &release,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups" +
					" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release = ? AND service = ? AND source = ?",
				args: []any{env, "value1", "value2", release, service, source},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetErrorGroupsRequest{},

			mockConn: &mockConnRow{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConnRow{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRow(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetErrorGroupsTotal(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				require.Equal(t, uint64(0), got)
			}
		})
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

	tests := []struct {
		name string

		req             types.GetErrorGroupsRequest
		wantGroupsCount int
		wantErr         bool

		isSharded   bool
		queryFilter map[string]string

		mockConn *mockConnRows
	}{
		{
			name: "ok_by_releases",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Release: &release,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit:  20,
				Offset: 5,
				Order:  types.OrderFrequent,
			},
			wantGroupsCount: 2,

			mockConn: &mockConnRows{
				query: fmt.Sprintf(
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
						" FROM error_groups"+
						" WHERE service = ? AND _group_hash IN (%s)"+
						" GROUP BY _group_hash, source"+
						" ORDER BY seen_total DESC",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING any(release) = ? AND count() = ?"+
						" ORDER BY countMerge(seen_total) DESC"+
						" LIMIT 20 OFFSET 5",
				),
				args: []any{
					service,
					service, release, 1,
				},

				rows: &mockRowsCount{
					count:        2,
					isScanStruct: true,
				},
			},
		},
		{
			name: "ok_by_timerange_relative",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit: 10,
				Order: types.OrderLatest,
			},
			wantGroupsCount: 2,

			mockConn: &mockConnRows{
				query: fmt.Sprintf(
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
						" FROM error_groups"+
						" WHERE service = ? AND _group_hash IN (%s)"+
						" GROUP BY _group_hash, source"+
						" ORDER BY last_seen_at DESC",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING minMerge(first_seen_at) >= ?"+
						" ORDER BY maxMerge(last_seen_at) DESC"+
						" LIMIT 10 OFFSET 0",
				),
				args: []any{
					service,
					service, timeDiff,
				},

				rows: &mockRowsCount{
					count:        2,
					isScanStruct: true,
				},
			},
		},
		{
			name: "ok_by_timerange_absolute",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
				Limit: 10,
				Order: types.OrderLatest,
			},
			wantGroupsCount: 2,

			mockConn: &mockConnRows{
				query: fmt.Sprintf(
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
						" FROM error_groups"+
						" WHERE service = ? AND _group_hash IN (%s)"+
						" GROUP BY _group_hash, source"+
						" ORDER BY last_seen_at DESC",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING (minMerge(first_seen_at) >= ? AND minMerge(first_seen_at) <= ?)"+
						" ORDER BY maxMerge(last_seen_at) DESC"+
						" LIMIT 10 OFFSET 0",
				),
				args: []any{
					service,
					service, timeDiff, fakeNow(),
				},

				rows: &mockRowsCount{
					count:        2,
					isScanStruct: true,
				},
			},
		},
		{
			name: "ok_full_filters_sharded",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Env:     &env,
				Source:  &source,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit:  10,
				Offset: 20,
				Order:  types.OrderOldest,
			},
			wantGroupsCount: 2,

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},
			isSharded: true,

			mockConn: &mockConnRows{
				query: fmt.Sprintf(
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
						" FROM error_groups"+
						" WHERE env = ? AND filter1 = ? AND filter2 = ? AND service = ? AND source = ? AND _group_hash GLOBAL IN (%s)"+
						" GROUP BY _group_hash, source"+
						" ORDER BY first_seen_at",

					"SELECT DISTINCT _group_hash"+
						" FROM error_groups"+
						" WHERE env = ? AND filter1 = ? AND filter2 = ? AND service = ? AND source = ?"+
						" GROUP BY _group_hash"+
						" HAVING minMerge(first_seen_at) >= ?"+
						" ORDER BY minMerge(first_seen_at)"+
						" LIMIT 10 OFFSET 20",
				),
				args: []any{
					env, "value1", "value2", service, source,
					env, "value1", "value2", service, source, timeDiff,
				},

				rows: &mockRowsCount{
					count:        2,
					isScanStruct: true,
				},
			},
		},
		{
			name: "ok_no_rows",

			req:             types.GetErrorGroupsRequest{},
			wantGroupsCount: 0,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					scanErr:      someErr,
					isScanStruct: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConn)
			repo := newRepo(mockedConn, tt.isSharded, tt.queryFilter, fakeNow)

			got, err := repo.GetNewErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantGroupsCount, len(got))
		})
	}
}

func TestGetNewErrorGroupsTotal(t *testing.T) {
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

	tests := []struct {
		name string

		req     types.GetErrorGroupsRequest
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRow
	}{
		{
			name: "ok_by_releases",

			req: types.GetErrorGroupsRequest{
				Service: service,
				Release: &release,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},

			mockConn: &mockConnRow{
				query: fmt.Sprintf(
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING any(release) = ? AND count() = ?",
				),
				args: []any{service, release, 1},
			},
		},
		{
			name: "ok_by_timerange_relative",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},

			mockConn: &mockConnRow{
				query: fmt.Sprintf(
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING minMerge(first_seen_at) >= ?",
				),
				args: []any{service, timeDiff},
			},
		},
		{
			name: "ok_by_timerange_absolute",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
			},

			mockConn: &mockConnRow{
				query: fmt.Sprintf(
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ?"+
						" GROUP BY _group_hash"+
						" HAVING (minMerge(first_seen_at) >= ? AND minMerge(first_seen_at) <= ?)",
				),
				args: []any{service, timeDiff, fakeNow()},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupsRequest{
				Service: service,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Env:    &env,
				Source: &source,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRow{
				query: fmt.Sprintf(
					"SELECT count() FROM (%s) AS subQ",

					"SELECT _group_hash"+
						" FROM error_groups"+
						" WHERE service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
						" GROUP BY _group_hash"+
						" HAVING minMerge(first_seen_at) >= ?",
				),
				args: []any{service, "value1", "value2", env, source, timeDiff},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetErrorGroupsRequest{},

			mockConn: &mockConnRow{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConnRow{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRow(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetNewErrorGroupsTotal(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				require.Equal(t, uint64(0), got)
			}
		})
	}
}

func TestGetTopErrorGroups(t *testing.T) {
	var (
		env    = "test-env"
		source = "test-source"

		fakeNow  = fakeNow(time.Now())
		duration = time.Hour * 24
		timeDiff = fakeNow().Add(-duration.Abs())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req             types.GetTopErrorGroupsRequest
		wantGroupsCount int
		wantErr         bool

		isSharded   bool
		queryFilter map[string]string

		mockConns []*mockConnRows
	}{
		{
			name: "ok_no_timerange",

			req: types.GetTopErrorGroupsRequest{
				Limit:  10,
				Offset: 20,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total"+
							" FROM error_groups"+
							" WHERE (1=1) AND _group_hash IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY seen_total DESC",

						"SELECT _group_hash"+
							" FROM error_groups_brief"+
							" WHERE (1=1)"+
							" GROUP BY _group_hash"+
							" ORDER BY countMerge(seen_total) DESC"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_relative",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
				Limit:  10,
				Offset: 20,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE (1=1) AND toStartOfHour(start_date) >= ?" +
						" GROUP BY _group_hash" +
						" ORDER BY count DESC" +
						" LIMIT 10 OFFSET 20",
					args: []any{timeDiff},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, source, any(message) as message" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?)" +
						" GROUP BY _group_hash, source",
					args: []any{uint64(123), uint64(456)},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_timerange_absolute",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
				Limit:  10,
				Offset: 20,
			},
			wantGroupsCount: 2,

			mockConns: []*mockConnRows{
				{
					query: "SELECT _group_hash, countMerge(counts) as count" +
						" FROM agg_events_10min" +
						" WHERE (1=1) AND (toStartOfHour(start_date) >= ? AND toStartOfHour(start_date) <= ?)" +
						" GROUP BY _group_hash" +
						" ORDER BY count DESC" +
						" LIMIT 10 OFFSET 20",
					args: []any{timeDiff, fakeNow()},

					rows: &mockRowsScanStruct{
						scanStructFns: []func(any) error{
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 123}
								return nil
							},
							func(v any) error {
								*v.(*errorCount) = errorCount{Hash: 456}
								return nil
							},
						},
					},
				},
				{
					query: "SELECT _group_hash, source, any(message) as message" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?)" +
						" GROUP BY _group_hash, source",
					args: []any{uint64(123), uint64(456)},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_full_filters_sharded",

			req: types.GetTopErrorGroupsRequest{
				Env:    &env,
				Source: &source,
				Limit:  10,
				Offset: 20,
			},
			wantGroupsCount: 2,

			isSharded: true,
			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConns: []*mockConnRows{
				{
					query: fmt.Sprintf(
						"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total"+
							" FROM error_groups"+
							" WHERE env = ? AND filter1 = ? AND filter2 = ? AND source = ? AND _group_hash GLOBAL IN (%s)"+
							" GROUP BY _group_hash, source"+
							" ORDER BY seen_total DESC",

						"SELECT DISTINCT _group_hash"+
							" FROM error_groups_brief"+
							" WHERE env = ? AND filter1 = ? AND filter2 = ? AND source = ?"+
							" GROUP BY _group_hash"+
							" ORDER BY countMerge(seen_total) DESC"+
							" LIMIT 10 OFFSET 20",
					),
					args: []any{
						env, "value1", "value2", source,
						env, "value1", "value2", source,
					},

					rows: &mockRowsCount{
						count:        2,
						isScanStruct: true,
					},
				},
			},
		},
		{
			name: "ok_no_rows_no_timerange",

			req:             types.GetTopErrorGroupsRequest{},
			wantGroupsCount: 0,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "ok_no_rows_timerange",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},
			wantGroupsCount: 0,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetTopErrorGroupsRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					err: someErr,
				},
			},
		},
		{
			name: "err_scan",

			req:     types.GetTopErrorGroupsRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						scanErr:      someErr,
						isScanStruct: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConns...)
			repo := newRepo(mockedConn, tt.isSharded, tt.queryFilter, fakeNow)

			got, err := repo.GetTopErrorGroups(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantGroupsCount, len(got))
		})
	}
}

func TestGetTopErrorGroupsTotal(t *testing.T) {
	var (
		env    = "test-env"
		source = "test-source"

		fakeNow  = fakeNow(time.Now())
		duration = time.Hour * 24
		timeDiff = fakeNow().Add(-duration.Abs())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req     types.GetTopErrorGroupsRequest
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRow
	}{
		{
			name: "ok_no_timerange",

			req: types.GetTopErrorGroupsRequest{},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups_brief" +
					" WHERE (1=1)",

				args: []any{},
			},
		},
		{
			name: "ok_timerange_relative",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM agg_events_10min" +
					" WHERE (1=1) AND toStartOfHour(start_date) >= ?",

				args: []any{timeDiff},
			},
		},
		{
			name: "ok_timerange_absolute",

			req: types.GetTopErrorGroupsRequest{
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
			},

			mockConn: &mockConnRow{
				query: "" +
					"SELECT uniq(_group_hash)" +
					" FROM agg_events_10min" +
					" WHERE (1=1) AND (toStartOfHour(start_date) >= ? AND toStartOfHour(start_date) <= ?)",

				args: []any{timeDiff, fakeNow()},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetTopErrorGroupsRequest{
				Env:    &env,
				Source: &source,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups_brief" +
					" WHERE env = ? AND filter1 = ? AND filter2 = ? AND source = ?",

				args: []any{env, "value1", "value2", source},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetTopErrorGroupsRequest{},

			mockConn: &mockConnRow{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.GetTopErrorGroupsRequest{},
			wantErr: true,

			mockConn: &mockConnRow{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRow(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetTopErrorGroupsTotal(context.Background(), tt.req)

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

	tests := []struct {
		name string

		req        types.DiffByReleasesRequest
		wantGroups []types.DiffGroup
		wantErr    bool

		queryFilter map[string]string

		mockConns []*mockConnRows
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

			mockConns: []*mockConnRows{
				{
					query: "" +
						"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
						" FROM error_groups" +
						" WHERE release IN (?,?) AND service = ?" +
						" GROUP BY _group_hash, source" +
						" ORDER BY countMerge(seen_total) DESC" +
						" LIMIT 20 OFFSET 0",
					args: []any{releases[0], releases[1], service},

					rows: &mockRowsScan{
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
				{
					query: "" +
						"SELECT _group_hash, release, countMerge(seen_total) as seen_total" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?) AND release IN (?,?) AND service = ?" +
						" GROUP BY _group_hash, release",
					args: []any{uint64(123), uint64(456), releases[0], releases[1], service},

					rows: &mockRowsScan{
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
			mockConns: []*mockConnRows{
				{
					query: "" +
						"SELECT _group_hash, source, any(message) as message, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at" +
						" FROM error_groups" +
						" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release IN (?,?) AND service = ? AND source = ?" +
						" GROUP BY _group_hash, source" +
						" ORDER BY last_seen_at DESC" +
						" LIMIT 20 OFFSET 5",
					args: []any{env, "value1", "value2", releases[0], releases[1], service, source},

					rows: &mockRowsScan{
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
				{
					query: "" +
						"SELECT _group_hash, release, countMerge(seen_total) as seen_total" +
						" FROM error_groups" +
						" WHERE _group_hash IN (?,?) AND env = ? AND filter1 = ? AND filter2 = ? AND release IN (?,?) AND service = ? AND source = ?" +
						" GROUP BY _group_hash, release",
					args: []any{uint64(123), uint64(456), env, "value1", "value2", releases[0], releases[1], service, source},

					rows: &mockRowsScan{
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
		},
		{
			name: "ok_no_rows",

			req:        types.DiffByReleasesRequest{},
			wantGroups: nil,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsCount{
						count: 0,
					},
				},
			},
		},
		{
			name: "err_query_groups",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{err: someErr},
			},
		},
		{
			name: "err_scan_groups",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsScan{
						scanFns: []func(...any) error{
							func(args ...any) error { return someErr },
						},
						scanErr: true,
					},
				},
			},
		},
		{
			name: "err_query",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsScan{
						scanFns: []func(...any) error{
							func(args ...any) error {
								*args[0].(*uint64) = 123
								return nil
							},
						},
					},
				},
				{
					err: someErr,
				},
			},
		},
		{
			name: "err_scan",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConns: []*mockConnRows{
				{
					rows: &mockRowsScan{
						scanFns: []func(...any) error{
							func(args ...any) error {
								*args[0].(*uint64) = 123
								return nil
							},
						},
					},
				},
				{
					rows: &mockRowsScan{
						scanFns: []func(...any) error{
							func(args ...any) error { return someErr },
						},
						scanErr: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConns...)
			repo := newRepo(mockedConn, true, tt.queryFilter, time.Now)

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

	tests := []struct {
		name string

		req     types.DiffByReleasesRequest
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRow
	}{
		{
			name: "ok",

			req: types.DiffByReleasesRequest{
				Service:  service,
				Releases: releases,
			},

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups" +
					" WHERE release IN (?,?) AND service = ?",
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

			mockConn: &mockConnRow{
				query: "SELECT uniq(_group_hash)" +
					" FROM error_groups" +
					" WHERE env = ? AND filter1 = ? AND filter2 = ? AND release IN (?,?) AND service = ? AND source = ?",
				args: []any{env, "value1", "value2", releases[0], releases[1], service, source},
			},
		},
		{
			name: "ok_no_rows",

			req: types.DiffByReleasesRequest{},

			mockConn: &mockConnRow{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.DiffByReleasesRequest{},
			wantErr: true,

			mockConn: &mockConnRow{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRow(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, time.Now)

			got, err := repo.DiffByReleasesTotal(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				require.Equal(t, uint64(0), got)
			}
		})
	}
}

func TestGetErrorHist(t *testing.T) {
	var (
		groupHash = uint64(123)
		service   = "test-svc"
		release   = "test-release"
		env       = "test-env"
		source    = "test-source"

		fakeNow  = fakeNow(time.Now())
		duration = time.Hour
		timeDiff = fakeNow().Add(-duration.Abs())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req              types.GetErrorHistRequest
		wantBucketsCount int
		wantErr          bool

		queryFilter map[string]string

		mockConn *mockConnRows
	}{
		{
			name: "ok",

			req:              types.GetErrorHistRequest{},
			wantBucketsCount: 2,

			mockConn: &mockConnRows{
				query: "" +
					"SELECT toStartOfMonth(start_date), countMerge(counts) as counts" +
					" FROM agg_events_1d" +
					" GROUP BY toStartOfMonth(start_date)" +
					" ORDER BY toStartOfMonth(start_date)",
				args: []any{},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_timerange_relative",

			req: types.GetErrorHistRequest{
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},
			wantBucketsCount: 2,

			mockConn: &mockConnRows{
				query: "" +
					"SELECT start_date, countMerge(counts) as counts" +
					" FROM agg_events_10min" +
					" WHERE start_date >= ?" +
					" GROUP BY start_date" +
					" ORDER BY start_date",
				args: []any{timeDiff},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_timerange_absolute",

			req: types.GetErrorHistRequest{
				TimeRange: &types.TimeRange{
					From: timeDiff,
					To:   fakeNow(),
				},
			},
			wantBucketsCount: 2,

			mockConn: &mockConnRows{
				query: "" +
					"SELECT start_date, countMerge(counts) as counts" +
					" FROM agg_events_10min" +
					" WHERE (start_date >= ? AND start_date <= ?)" +
					" GROUP BY start_date" +
					" ORDER BY start_date",
				args: []any{timeDiff, fakeNow()},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorHistRequest{
				GroupHash: &groupHash,
				Service:   &service,
				Env:       &env,
				Source:    &source,
				Release:   &release,
				TimeRange: &types.TimeRange{
					Duration: duration,
				},
			},
			wantBucketsCount: 2,

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRows{
				query: "" +
					"SELECT start_date, countMerge(counts) as counts" +
					" FROM agg_events_10min" +
					" WHERE filter1 = ? AND filter2 = ? AND _group_hash = ? AND env = ? AND source = ? AND service = ? AND release = ? AND start_date >= ?" +
					" GROUP BY start_date" +
					" ORDER BY start_date",
				args: []any{"value1", "value2", groupHash, env, source, service, release, timeDiff},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_no_rows",

			req:              types.GetErrorHistRequest{},
			wantBucketsCount: 0,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetErrorHistRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorHistRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					scanErr: someErr,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetErrorHist(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantBucketsCount, len(got.Buckets))
		})
	}
}

func TestGetErrorDetails(t *testing.T) {
	var (
		groupHash = uint64(123)
		service   = "test-svc"
		release   = "test-release"
		env       = "test-env"
		source    = "test-source"

		fakeNow = fakeNow(time.Now())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req     types.GetErrorGroupDetailsRequest
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRow
	}{
		{
			name: "ok",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: groupHash,
			},

			mockConn: &mockConnRow{
				query: "" +
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at, max(log_tags) as log_tags" +
					" FROM error_groups" +
					" WHERE _group_hash = ?" +
					" GROUP BY _group_hash, source",
				args: []any{groupHash},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: groupHash,
				Env:       &env,
				Source:    &source,
				Service:   &service,
				Release:   &release,
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRow{
				query: "" +
					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at, max(log_tags) as log_tags" +
					" FROM error_groups" +
					" WHERE _group_hash = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ? AND service = ? AND release = ?" +
					" GROUP BY _group_hash, source",
				args: []any{groupHash, "value1", "value2", env, source, service, release},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetErrorGroupDetailsRequest{},

			mockConn: &mockConnRow{
				scanErr: sql.ErrNoRows,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupDetailsRequest{},
			wantErr: true,

			mockConn: &mockConnRow{
				scanErr: someErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRow(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			_, err := repo.GetErrorDetails(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGetErrorCounts(t *testing.T) {
	var (
		groupHash = uint64(123)
		service   = "test-svc"
		release   = "test-release"
		env       = "test-env"
		source    = "test-source"

		fakeNow = fakeNow(time.Now())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req     types.GetErrorGroupDetailsRequest
		want    types.ErrorGroupCounts
		wantErr bool

		queryFilter map[string]string

		mockConn *mockConnRows
	}{
		{
			name: "ok",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: groupHash,
			},
			want: types.ErrorGroupCounts{
				ByEnv: types.ErrorGroupCount{
					"env1": 10,
					"env2": 20,
				},
				BySource: types.ErrorGroupCount{
					"source1": 10,
					"source2": 20,
				},
				ByService: types.ErrorGroupCount{
					"service1": 30,
				},
				ByRelease: types.ErrorGroupCount{},
			},

			mockConn: &mockConnRows{
				query: "" +
					"SELECT countMerge(seen_total) as count, env, source, service" +
					" FROM error_groups" +
					" WHERE _group_hash = ?" +
					" GROUP BY env, source, service",
				args: []any{groupHash},

				rows: &mockRowsScanStruct{
					scanStructFns: []func(any) error{
						func(ec any) error {
							*ec.(*errDetailsCount) = errDetailsCount{
								Count:   10,
								Env:     "env1",
								Source:  "source1",
								Service: "service1",
							}
							return nil
						},
						func(ec any) error {
							*ec.(*errDetailsCount) = errDetailsCount{
								Count:   20,
								Env:     "env2",
								Source:  "source2",
								Service: "service1",
							}
							return nil
						},
					},
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetErrorGroupDetailsRequest{
				GroupHash: groupHash,
				Env:       &env,
				Source:    &source,
				Service:   &service,
				Release:   &release,
			},
			want: types.ErrorGroupCounts{
				ByEnv:     types.ErrorGroupCount{env: 10},
				BySource:  types.ErrorGroupCount{source: 10},
				ByService: types.ErrorGroupCount{service: 10},
				ByRelease: types.ErrorGroupCount{release: 10},
			},

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRows{
				query: "" +
					"SELECT countMerge(seen_total) as count, env, source, service, release" +
					" FROM error_groups" +
					" WHERE _group_hash = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ? AND service = ? AND release = ?" +
					" GROUP BY env, source, service, release",
				args: []any{groupHash, "value1", "value2", env, source, service, release},

				rows: &mockRowsScanStruct{
					scanStructFns: []func(any) error{
						func(ec any) error {
							*ec.(*errDetailsCount) = errDetailsCount{
								Count:   10,
								Env:     env,
								Source:  source,
								Service: service,
								Release: release,
							}
							return nil
						},
					},
				},
			},
		},
		{
			name: "ok_no_rows",

			req: types.GetErrorGroupDetailsRequest{},
			want: types.ErrorGroupCounts{
				ByEnv:     types.ErrorGroupCount{},
				BySource:  types.ErrorGroupCount{},
				ByService: types.ErrorGroupCount{},
				ByRelease: types.ErrorGroupCount{},
			},

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetErrorGroupDetailsRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetErrorGroupDetailsRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				rows: &mockRowsScanStruct{
					scanStructFns: []func(any) error{
						func(_ any) error {
							return someErr
						},
					},
					scanErr: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetErrorCounts(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetServices(t *testing.T) {
	var (
		query = "test"
		env   = "test-env"

		fakeNow = fakeNow(time.Now())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req               types.GetServicesRequest
		wantServicesCount int
		wantErr           bool

		queryFilter map[string]string

		mockConn *mockConnRows
	}{
		{
			name: "ok",

			req:               types.GetServicesRequest{},
			wantServicesCount: 2,

			mockConn: &mockConnRows{
				query: "" +
					"SELECT DISTINCT service" +
					" FROM services" +
					" WHERE service <> ?" +
					" ORDER BY service",
				args: []any{""},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetServicesRequest{
				Query:  query,
				Env:    &env,
				Limit:  5,
				Offset: 10,
			},
			wantServicesCount: 5,

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRows{
				query: "" +
					"SELECT DISTINCT service" +
					" FROM services" +
					" WHERE service <> ? AND filter1 = ? AND filter2 = ? AND startsWith(service, ?) AND env = ?" +
					" ORDER BY service" +
					" LIMIT 5 OFFSET 10",
				args: []any{"", "value1", "value2", query, env},

				rows: &mockRowsCount{
					count: 5,
				},
			},
		},
		{
			name: "ok_no_rows",

			req:               types.GetServicesRequest{},
			wantServicesCount: 0,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetServicesRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetServicesRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					scanErr: someErr,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetServices(context.Background(), tt.req)

			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantServicesCount, len(got))
		})
	}
}

func TestGetReleases(t *testing.T) {
	var (
		service = "test-service"
		env     = "test-env"

		fakeNow = fakeNow(time.Now())

		someErr = errors.New("some err")
	)

	tests := []struct {
		name string

		req               types.GetReleasesRequest
		wantReleasesCount int
		wantErr           bool

		queryFilter map[string]string

		mockConn *mockConnRows
	}{
		{
			name: "ok",

			req: types.GetReleasesRequest{
				Service: service,
			},
			wantReleasesCount: 2,

			mockConn: &mockConnRows{
				query: "" +
					"SELECT DISTINCT release" +
					" FROM services" +
					" WHERE (service = ? AND release <> ?)" +
					" ORDER BY ttl DESC",
				args: []any{service, ""},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_full_filters",

			req: types.GetReleasesRequest{
				Service: service,
				Env:     &env,
			},
			wantReleasesCount: 2,

			queryFilter: map[string]string{
				"filter1": "value1",
				"filter2": "value2",
			},

			mockConn: &mockConnRows{
				query: "" +
					"SELECT DISTINCT release" +
					" FROM services" +
					" WHERE (service = ? AND release <> ?) AND filter1 = ? AND filter2 = ? AND env = ?" +
					" ORDER BY ttl DESC",
				args: []any{service, "", "value1", "value2", env},

				rows: &mockRowsCount{
					count: 2,
				},
			},
		},
		{
			name: "ok_no_rows",

			req:               types.GetReleasesRequest{},
			wantReleasesCount: 0,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					count: 0,
				},
			},
		},
		{
			name: "err_query",

			req:     types.GetReleasesRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				err: someErr,
			},
		},
		{
			name: "err_scan",

			req:     types.GetReleasesRequest{},
			wantErr: true,

			mockConn: &mockConnRows{
				rows: &mockRowsCount{
					scanErr: someErr,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockedConn := initMockConnRows(t, tt.mockConn)
			repo := newRepo(mockedConn, true, tt.queryFilter, fakeNow)

			got, err := repo.GetReleases(context.Background(), tt.req)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.wantReleasesCount, len(got))
		})
	}
}
