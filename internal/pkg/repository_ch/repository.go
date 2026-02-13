package repositorych

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"

	"github.com/ozontech/seq-ui/internal/app/types"
	sqlb "github.com/ozontech/seq-ui/internal/pkg/repository/sql_builder"
)

type Repository interface {
	GetErrorGroups(context.Context, types.GetErrorGroupsRequest) ([]types.ErrorGroup, error)
	GetErrorGroupsCount(context.Context, types.GetErrorGroupsRequest) (uint64, error)
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
}

func New(conn driver.Conn, sharded bool, queryFilter map[string]string) Repository {
	return &repository{
		conn:        newConn(conn),
		sharded:     sharded,
		queryFilter: queryFilter,
	}
}

func (r *repository) GetErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, error) {
	// we need this subquery to make query faster, see https://github.com/ClickHouse/ClickHouse/issues/7187
	subQ := sqlb.
		Select("_group_hash").
		From("error_groups").
		Where(sq.Eq{"service": req.Service}).
		GroupBy("_group_hash", "service").
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	if r.sharded {
		subQ = subQ.Distinct()
	}

	for col, val := range r.queryFilter {
		subQ = subQ.Where(sq.Eq{col: val}).GroupBy(col)
	}

	if req.Env != nil && *req.Env != "" {
		subQ = subQ.Where(sq.Eq{"env": req.Env}).GroupBy("env")
	}
	if req.Release != nil && *req.Release != "" {
		subQ = subQ.Where(sq.Eq{"release": req.Release}).GroupBy("release")
	}
	if req.Duration != nil && *req.Duration != 0 {
		subQ = subQ.Having(sq.GtOrEq{"maxMerge(last_seen_at)": time.Now().Add(-req.Duration.Abs())})
	}
	if req.Source != nil && *req.Source != "" {
		subQ = subQ.Where(sq.Eq{"source": req.Source}).GroupBy("source")
	}
	subQ = orderBy(subQ, req.Order, true)

	subQuery, subArgs := subQ.MustSql()

	in := "IN"
	if r.sharded {
		in = "GLOBAL IN"
	}
	q := sqlb.
		Select(
			"_group_hash as group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		).
		From("error_groups").
		Where(fmt.Sprintf("_group_hash %s (%s)", in, subQuery), subArgs...).
		GroupBy("_group_hash", "service", "source")

	// using string formatting below because squirrel doesn't support subquery in WHERE clause
	q = q.Where(fmt.Sprintf("service = '%s'", req.Service))

	for col, val := range r.queryFilter {
		q = q.Where(fmt.Sprintf("%s = '%s'", col, val)).GroupBy(col)
	}

	if req.Source != nil && *req.Source != "" {
		q = q.Where(fmt.Sprintf("source = '%s'", *req.Source))
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(fmt.Sprintf("env = '%s'", *req.Env)).GroupBy("env")
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(fmt.Sprintf("release = '%s'", *req.Release)).GroupBy("release")
	}
	q = orderBy(q, req.Order, false)

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups: %w", err)
	}

	var errorGroups []types.ErrorGroup
	for rows.Next() {
		var group types.ErrorGroup
		err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.SeenTotal,
			&group.FirstSeenAt,
			&group.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		errorGroups = append(errorGroups, group)
	}

	return errorGroups, nil
}

func (r *repository) GetErrorGroupsCount(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) (uint64, error) {
	subQ := sqlb.
		Select("maxMerge(last_seen_at) AS last_seen_at").
		From("error_groups").
		Where(sq.Eq{"service": req.Service}).
		GroupBy("_group_hash", "service")

	for col, val := range r.queryFilter {
		subQ = subQ.Where(sq.Eq{col: val}).GroupBy(col)
	}

	if req.Env != nil && *req.Env != "" {
		subQ = subQ.Where(sq.Eq{"env": req.Env}).GroupBy("env")
	}
	if req.Release != nil && *req.Release != "" {
		subQ = subQ.Where(sq.Eq{"release": req.Release}).GroupBy("release")
	}
	if req.Duration != nil && *req.Duration != 0 {
		subQ = subQ.Having(sq.GtOrEq{"last_seen_at": time.Now().Add(-req.Duration.Abs())})
	}
	if req.Source != nil && *req.Source != "" {
		subQ = subQ.Where(sq.Eq{"source": req.Source}).GroupBy("source")
	}

	q := sqlb.Select("count()").FromSelect(subQ, "subQ")

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	row := r.conn.QueryRow(ctx, metricLabels, query, args...)

	var total uint64
	if err := row.Scan(&total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to get error groups count: %w", err)
	}

	return total, nil
}

func (r *repository) GetErrorHist(
	ctx context.Context,
	req types.GetErrorHistRequest,
) ([]types.ErrorHistBucket, error) {
	startDate := getHistBucketSize(req.Duration)

	q := sqlb.
		Select(
			startDate,
			"countMerge(counts) as counts",
		).
		From("agg_events_10min").
		Where(sq.Eq{"service": req.Service}).
		GroupBy(startDate, "service").
		OrderBy(startDate)

	for col, val := range r.queryFilter {
		q = q.Where(sq.Eq{col: val}).GroupBy(col)
	}

	if req.GroupHash != nil && *req.GroupHash != 0 {
		q = q.Where(sq.Eq{"_group_hash": req.GroupHash}).GroupBy("_group_hash")
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": req.Env}).GroupBy("env")
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(sq.Eq{"release": req.Release}).GroupBy("release")
	}
	if req.Duration != nil && *req.Duration != 0 {
		q = q.Where(sq.GtOrEq{startDate: time.Now().Add(-req.Duration.Abs())})
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": req.Source}).GroupBy("source")
	}

	query, args := q.MustSql()
	metricLabels := []string{"agg_events_10min", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error hist: %w", err)
	}

	var buckets []types.ErrorHistBucket
	for rows.Next() {
		var bucket types.ErrorHistBucket
		if err := rows.Scan(&bucket.Time, &bucket.Count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (r *repository) GetErrorDetails(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupDetails, error) {
	q := sqlb.
		Select(
			"_group_hash as group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
			"max(log_tags) as log_tags",
		).
		From("error_groups").
		Where(sq.Eq{
			"service":     req.Service,
			"_group_hash": req.GroupHash,
		}).
		GroupBy("_group_hash", "service", "source")

	for col, val := range r.queryFilter {
		q = q.Where(sq.Eq{col: val}).GroupBy(col)
	}

	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": req.Env}).GroupBy("env")
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(sq.Eq{"release": req.Release}).GroupBy("release")
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": req.Source})
	}

	var details types.ErrorGroupDetails

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	row := r.conn.QueryRow(ctx, metricLabels, query, args...)
	err := row.Scan(
		&details.GroupHash,
		&details.Source,
		&details.Message,
		&details.SeenTotal,
		&details.FirstSeenAt,
		&details.LastSeenAt,
		&details.LogTags,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		incErrorMetric(err, metricLabels)
		return details, fmt.Errorf("failed to get error details: %w", err)
	}

	return details, nil
}

func (r *repository) GetErrorCounts(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupCounts, error) {
	counts := types.ErrorGroupCounts{
		ByEnv:     types.ErrorGroupCount{},
		ByRelease: types.ErrorGroupCount{},
	}

	q := sqlb.
		Select("countMerge(seen_total) as seen_total", "env", "release").
		From("error_groups").
		Where(sq.Eq{
			"service":     req.Service,
			"_group_hash": req.GroupHash,
		}).
		GroupBy("_group_hash", "service", "env", "release").
		OrderBy("seen_total DESC")

	for col, val := range r.queryFilter {
		q = q.Where(sq.Eq{col: val}).GroupBy(col)
	}

	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(sq.Eq{"release": *req.Release})
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": *req.Source})
	}

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return counts, fmt.Errorf("failed to get error counts: %w", err)
	}

	for rows.Next() {
		var (
			seen         uint64
			env, release string
		)
		if err := rows.Scan(&seen, &env, &release); err != nil {
			return counts, fmt.Errorf("failed to scan row: %w", err)
		}
		counts.ByEnv[env] += seen
		counts.ByRelease[release] += seen
	}

	return counts, nil
}

func (r *repository) GetErrorReleases(
	ctx context.Context,
	req types.GetErrorGroupReleasesRequest,
) ([]string, error) {
	q := sqlb.
		Select("release").Distinct().
		From("services").
		Where(sq.And{
			sq.Eq{"service": req.Service},
			sq.NotEq{"release": ""},
		}).
		OrderBy("release")

	for col, val := range r.queryFilter {
		q = q.Where(sq.Eq{col: val})
	}

	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": req.Env})
	}

	query, args := q.MustSql()
	metricLabels := []string{"services", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}

	releases := make([]string, 0)
	for rows.Next() {
		var release string
		if err := rows.Scan(&release); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		if release == "" {
			continue
		}
		releases = append(releases, release)
	}

	return releases, nil
}

func (r *repository) GetServices(
	ctx context.Context,
	req types.GetServicesRequest,
) ([]string, error) {
	q := sqlb.
		Select("service").Distinct().
		From("services").
		Where("startsWith(service, ?)", req.Query).
		Where(sq.NotEq{"service": ""}).
		OrderBy("service").
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	for col, val := range r.queryFilter {
		q = q.Where(sq.Eq{col: val})
	}

	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": req.Env})
	}

	query, args := q.MustSql()
	metricLabels := []string{"services", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := make([]string, 0)
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		services = append(services, service)
	}

	return services, nil
}

func (r *repository) DiffByReleases(
	ctx context.Context,
	req types.DiffByReleasesRequest,
) ([]types.DiffGroup, error) {
	where := sq.Eq{
		"service": req.Service,
		"release": req.Releases,
	}
	for col, val := range r.queryFilter {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	groupsQ := sqlb.
		Select(
			"_group_hash",
			"source",
			"any(message) as message",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		).
		From("error_groups").
		Where(where).
		GroupBy("_group_hash", "source").
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	switch req.Order {
	case types.OrderFrequent:
		groupsQ = groupsQ.OrderBy("countMerge(seen_total) DESC")
	case types.OrderLatest:
		groupsQ = groupsQ.OrderBy("last_seen_at DESC")
	case types.OrderOldest:
		groupsQ = groupsQ.OrderBy("first_seen_at")
	}

	groupsQuery, args := groupsQ.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, groupsQuery, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups: %w", err)
	}

	var (
		diffGroups []types.DiffGroup
		idxByHash  = map[uint64]int{}
	)
	for rows.Next() {
		group := types.DiffGroup{
			ReleaseInfos: make(map[string]types.DiffReleaseInfo),
		}

		err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.FirstSeenAt,
			&group.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		diffGroups = append(diffGroups, group)
		idxByHash[group.Hash] = len(diffGroups) - 1
	}

	where["_group_hash"] = slices.Collect(maps.Keys(idxByHash))

	q := sqlb.
		Select(
			"_group_hash",
			"release",
			"countMerge(seen_total) as seen_total",
		).
		From("error_groups").
		Where(where).
		GroupBy("_group_hash", "release")

	query, args := q.MustSql()
	rows, err = r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups by release: %w", err)
	}

	for rows.Next() {
		var (
			hash, seenTotal uint64
			release         string
		)
		if err := rows.Scan(&hash, &release, &seenTotal); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if idx, ok := idxByHash[hash]; ok {
			diffGroups[idx].ReleaseInfos[release] = types.DiffReleaseInfo{
				SeenTotal: seenTotal,
			}
		}
	}

	return diffGroups, nil
}

func (r *repository) DiffByReleasesTotal(
	ctx context.Context,
	req types.DiffByReleasesRequest,
) (uint64, error) {
	subQ := sqlb.
		Select("_group_hash").
		From("error_groups").
		Where(sq.Eq{
			"service": req.Service,
			"release": req.Releases,
		}).
		GroupBy("_group_hash")

	for col, val := range r.queryFilter {
		subQ = subQ.Where(sq.Eq{col: val})
	}
	if req.Env != nil && *req.Env != "" {
		subQ = subQ.Where(sq.Eq{"env": *req.Env})
	}
	if req.Source != nil && *req.Source != "" {
		subQ = subQ.Where(sq.Eq{"source": *req.Source})
	}

	q := sqlb.Select("count()").FromSelect(subQ, "subQ")

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	row := r.conn.QueryRow(ctx, metricLabels, query, args...)

	var total uint64
	if err := row.Scan(&total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to get error groups count: %w", err)
	}

	return total, nil
}

func orderBy(q sq.SelectBuilder, o types.ErrorGroupsOrder, sub bool) sq.SelectBuilder {
	seenTotal := "seen_total DESC"
	lastSeenAt := "last_seen_at DESC"
	firstSeenAt := "first_seen_at"
	if sub {
		seenTotal = "countMerge(seen_total) DESC"
		lastSeenAt = "maxMerge(last_seen_at) DESC"
		firstSeenAt = "minMerge(first_seen_at)"
	}

	switch o {
	case types.OrderFrequent:
		q = q.OrderBy(seenTotal)
	case types.OrderLatest:
		q = q.OrderBy(lastSeenAt)
	case types.OrderOldest:
		q = q.OrderBy(firstSeenAt)
	}
	return q
}

func getHistBucketSize(d *time.Duration) string {
	const (
		startDate   = "start_date"
		startOfHour = "toStartOfHour(start_date)"
		startOfDay  = "toStartOfDay(start_date)"
		day         = 24 * time.Hour
	)

	if d == nil {
		return startOfDay
	}

	duration := *d
	switch {
	case duration < 7*time.Hour:
		return startDate
	case duration < 7*day:
		return startOfHour
	case duration >= 7*day:
		return startOfDay
	default:
		return startOfDay
	}
}
