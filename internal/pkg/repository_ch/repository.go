package repositorych

import (
	"context"
	"iter"
	"maps"
	"slices"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/ozontech/seq-ui/internal/app/types"
)

type Repository interface {
	GetErrorGroups(context.Context, types.GetErrorGroupsRequest) ([]types.ErrorGroup, error)
	GetErrorGroupsCount(context.Context, types.GetErrorGroupsRequest) (uint64, error)
	GetNewErrorGroups(context.Context, types.GetErrorGroupsRequest) ([]types.ErrorGroup, error)
	GetNewErrorGroupsCount(context.Context, types.GetErrorGroupsRequest) (uint64, error)
	GetErrorHist(context.Context, types.GetErrorHistRequest) ([]types.ErrorHistBucket, error)
	GetErrorDetails(context.Context, types.GetErrorGroupDetailsRequest) (types.ErrorGroupDetails, error)
	GetErrorCounts(context.Context, types.GetErrorGroupDetailsRequest) (types.ErrorGroupCounts, error)
	GetErrorReleases(context.Context, types.GetErrorGroupReleasesRequest) ([]string, error)
	GetServices(context.Context, types.GetServicesRequest) ([]string, error)
	DiffByReleases(context.Context, types.DiffByReleasesRequest) ([]types.DiffGroup, error)
	DiffByReleasesTotal(context.Context, types.DiffByReleasesRequest) (uint64, error)
}

type repository struct {
	conn    *conn
	sharded bool

	queryFilter map[string]string

	nowFn func() time.Time // for testing
}

func New(conn driver.Conn, sharded bool, queryFilter map[string]string) Repository {
	return newRepo(conn, sharded, queryFilter, time.Now)
}

func newRepo(conn driver.Conn, sharded bool, queryFilter map[string]string, nowFn func() time.Time) *repository {
	return &repository{
		conn:        newConn(conn),
		sharded:     sharded,
		queryFilter: queryFilter,
		nowFn:       nowFn,
	}
}

func (r *repository) queryFilters() iter.Seq2[string, string] {
	keys := slices.Collect(maps.Keys(r.queryFilter))
	slices.Sort(keys)
	return func(yield func(string, string) bool) {
		for _, key := range keys {
			if !yield(key, r.queryFilter[key]) {
				return
			}
		}
	}
}
