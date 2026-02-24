package repositorych

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/ozontech/seq-ui/internal/app/types"
// 	mock "github.com/ozontech/seq-ui/internal/pkg/repository_ch/mock"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/mock/gomock"
// )

// func fakeNow(now time.Time) func() time.Time {
// 	return func() time.Time {
// 		return now
// 	}
// }

// func TestGetNewErrorGroups(t *testing.T) {
// 	var (
// 		service = "test-svc"
// 		release = "test-release"
// 		env     = "test-env"
// 		source  = "test-source"

// 		fakeNow  = fakeNow(time.Now())
// 		duration = time.Hour * 24
// 		timeDiff = fakeNow().Add(-duration.Abs())

// 		someErr = errors.New("some err")
// 	)

// 	type mockRows struct {
// 		count   int
// 		scanErr error
// 	}

// 	type mockConn struct {
// 		query string
// 		args  []any

// 		rows *mockRows
// 		err  error
// 	}

// 	tests := []struct {
// 		name string

// 		req        types.GetErrorGroupsRequest
// 		wantGroups int
// 		wantErr    bool

// 		isSharded   bool
// 		queryFilter map[string]string
// 		mockConn    *mockConn
// 	}{
// 		{
// 			name: "ok_by_releases",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Release:  &release,
// 				Duration: &duration,
// 				Limit:    20,
// 				Offset:   5,
// 				Order:    types.OrderFrequent,
// 			},
// 			wantGroups: 2,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING any(release) = ? AND count() = ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 20 OFFSET 5",
// 				),
// 				args: []any{service, release, 1},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_by_duration",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Order:    types.OrderLatest,
// 			},
// 			wantGroups: 2,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY last_seen_at DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY maxMerge(last_seen_at) DESC"+
// 						" LIMIT 10 OFFSET 0",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_full_filters",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Env:      &env,
// 				Source:   &source,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderOldest,
// 			},

// 			wantGroups: 2,

// 			queryFilter: map[string]string{
// 				"filter1": "value1",
// 				"filter2": "value2",
// 			},
// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc' AND filter1 = 'value1' AND filter2 = 'value2' AND source = 'test-source' AND env = 'test-env'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY first_seen_at",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY minMerge(first_seen_at)"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, "value1", "value2", env, source, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_sharded",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantGroups: 2,

// 			isSharded: true,
// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash GLOBAL IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT DISTINCT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_no_rows",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantGroups: 0,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 0,
// 				},
// 			},
// 		},
// 		{
// 			name: "err_query",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantErr: true,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				err: someErr,
// 			},
// 		},
// 		{
// 			name: "err_scan",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantErr: true,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					scanErr: someErr,
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			ctrl := gomock.NewController(t)
// 			mockedConn := mock.NewMockConn(ctrl)

// 			repo := newRepo(mockedConn, tt.isSharded, tt.queryFilter, fakeNow)

// 			if tt.mockConn != nil {
// 				mockedRows := mock.NewMockRows(ctrl)
// 				if rows := tt.mockConn.rows; rows != nil {
// 					times := rows.count
// 					if rows.scanErr != nil {
// 						times = 1
// 					}

// 					mockedRows.EXPECT().Next().Return(true).Times(times)
// 					mockedRows.EXPECT().Scan(gomock.Any()).Return(rows.scanErr).Times(times)
// 					if rows.scanErr == nil {
// 						mockedRows.EXPECT().Next().Return(false).Times(1)
// 					}
// 				}

// 				mockedConn.EXPECT().Query(gomock.Any(), tt.mockConn.query, tt.mockConn.args...).Return(mockedRows, tt.mockConn.err).Times(1)
// 			}

// 			got, err := repo.GetNewErrorGroups(context.Background(), tt.req)
// 			require.Equal(t, tt.wantErr, err != nil)
// 			require.Equal(t, tt.wantGroups, len(got))
// 		})
// 	}
// }

// func TestGetNewErrorGroupsCount(t *testing.T) {
// 	const isSharded = true

// 	var (
// 		service = "test-svc"
// 		release = "test-release"
// 		//env     = "test-env"
// 		//source  = "test-source"

// 		fakeNow  = fakeNow(time.Now())
// 		duration = time.Hour * 24
// 		//timeDiff = fakeNow().Add(-duration.Abs())

// 		//someErr = errors.New("some err")
// 	)

// 	type mockConn struct {
// 		query string
// 		args  []any

// 		scanErr error
// 	}

// 	tests := []struct {
// 		name string

// 		req     types.GetErrorGroupsRequest
// 		wantErr bool

// 		queryFilter map[string]string
// 		mockConn    *mockConn
// 	}{
// 		{
// 			name: "ok_by_releases",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Release:  &release,
// 				Duration: &duration,
// 			},

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT count() FROM (%s) AS subQ",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING any(release) = ? AND count() = ?",
// 				),
// 				args: []any{service, release, 1},
// 			},
// 		},
// 		{
// 			name: "ok_by_duration",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Order:    types.OrderLatest,
// 			},
// 			wantGroups: 2,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY last_seen_at DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY maxMerge(last_seen_at) DESC"+
// 						" LIMIT 10 OFFSET 0",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_full_filters",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Env:      &env,
// 				Source:   &source,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderOldest,
// 			},

// 			wantGroups: 2,

// 			queryFilter: map[string]string{
// 				"filter1": "value1",
// 				"filter2": "value2",
// 			},
// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc' AND filter1 = 'value1' AND filter2 = 'value2' AND source = 'test-source' AND env = 'test-env'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY first_seen_at",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ? AND filter1 = ? AND filter2 = ? AND env = ? AND source = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY minMerge(first_seen_at)"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, "value1", "value2", env, source, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_sharded",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantGroups: 2,

// 			isSharded: true,
// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash GLOBAL IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT DISTINCT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 2,
// 				},
// 			},
// 		},
// 		{
// 			name: "ok_no_rows",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantGroups: 0,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					count: 0,
// 				},
// 			},
// 		},
// 		{
// 			name: "err_query",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantErr: true,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				err: someErr,
// 			},
// 		},
// 		{
// 			name: "err_scan",

// 			req: types.GetErrorGroupsRequest{
// 				Service:  service,
// 				Duration: &duration,
// 				Limit:    10,
// 				Offset:   20,
// 				Order:    types.OrderFrequent,
// 			},

// 			wantErr: true,

// 			mockConn: &mockConn{
// 				query: fmt.Sprintf(""+
// 					"SELECT _group_hash, source, any(message) as message, countMerge(seen_total) as seen_total, minMerge(first_seen_at) as first_seen_at, maxMerge(last_seen_at) as last_seen_at"+
// 					" FROM error_groups"+
// 					" WHERE _group_hash IN (%s) AND service = 'test-svc'"+
// 					" GROUP BY _group_hash, source"+
// 					" ORDER BY seen_total DESC",

// 					"SELECT _group_hash"+
// 						" FROM error_groups"+
// 						" WHERE service = ?"+
// 						" GROUP BY _group_hash, source"+
// 						" HAVING minMerge(first_seen_at) >= ?"+
// 						" ORDER BY countMerge(seen_total) DESC"+
// 						" LIMIT 10 OFFSET 20",
// 				),
// 				args: []any{service, timeDiff},

// 				rows: &mockRows{
// 					scanErr: someErr,
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			ctrl := gomock.NewController(t)
// 			mockedConn := mock.NewMockConn(ctrl)

// 			repo := newRepo(mockedConn, isSharded, tt.queryFilter, fakeNow)

// 			if tt.mockConn != nil {
// 				mockedRow := mock.NewMockRow(ctrl)
// 				mockedRow.EXPECT().Scan(gomock.Any()).Return(tt.mockConn.scanErr)

// 				mockedConn.EXPECT().QueryRow(gomock.Any(), tt.mockConn.query, tt.mockConn.args...).Return(mockedRow).Times(1)
// 			}

// 			_, err := repo.GetNewErrorGroupsCount(context.Background(), tt.req)
// 			require.Equal(t, tt.wantErr, err != nil)
// 		})
// 	}
// }

// func TestDiffByReleases(t *testing.T) {
// 	// Test implementation
// }

// func TestDiffByReleasesTotal(t *testing.T) {
// 	// Test implementation
// }
