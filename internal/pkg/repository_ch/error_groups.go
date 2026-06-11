// nolint:goconst
package repositorych

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/n-r-w/squirrel"

	"github.com/ozontech/seq-ui/internal/app/types"
)

func (r *repository) GetErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, error) {
	where := sq.Eq{
		"service": req.Service,
	}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}
	if req.Release != nil && *req.Release != "" {
		where["release"] = *req.Release
	}

	if req.Duration != nil && *req.Duration != 0 {
		var (
			hashes []uint64 // ordered
			infos  errorInfos
			counts errorCounts

			err error
		)

		if req.Order == types.OrderFrequent {
			counts, err = r.getErrorCounts(ctx, getErrorCountsParams{
				duration: req.Duration,
				where:    where,
				orderBy:  "count DESC",
				limit:    uint64(req.Limit),
				offset:   uint64(req.Offset),
			})
			if err != nil {
				return nil, err
			}

			if len(counts) == 0 {
				return nil, nil
			}

			hashes = counts.hashes()

			where["_group_hash"] = hashes
			infos, err = r.getErrorInfos(ctx, getErrorInfosParams{
				columns: []string{
					"_group_hash",
					"source",
					"any(message) as message",
					"minMerge(first_seen_at) as first_seen_at",
					"maxMerge(last_seen_at) as last_seen_at",
				},
				where: where,
			})
			if err != nil {
				return nil, err
			}
		} else {
			subQuery := r.getHashSubQuery(getHashSubQueryParams{
				table:    "error_groups",
				where:    where,
				duration: req.Duration,
				orderBy:  orderBy(req.Order, true),
				limit:    uint64(req.Limit),
				offset:   uint64(req.Offset),
			})

			infos, err := r.getErrorInfos(ctx, getErrorInfosParams{
				columns: []string{
					"_group_hash",
					"source",
					"any(message) as message",
					"minMerge(first_seen_at) as first_seen_at",
					"maxMerge(last_seen_at) as last_seen_at",
				},
				where:    where,
				subQuery: &subQuery,
				orderBy:  orderBy(req.Order, false),
			})
			if err != nil {
				return nil, err
			}

			if len(infos) == 0 {
				return nil, nil
			}

			hashes = infos.hashes()

			where["_group_hash"] = hashes
			counts, err = r.getErrorCounts(ctx, getErrorCountsParams{
				duration: req.Duration,
				where:    where,
			})
			if err != nil {
				return nil, err
			}
		}

		infoByHash := infos.mapByHash()
		countByHash := counts.mapByHash()

		var groups []types.ErrorGroup
		for _, hash := range hashes {
			info := infoByHash[hash]
			groups = append(groups, types.ErrorGroup{
				Hash:        hash,
				Source:      info.Source,
				Message:     info.Message,
				Count:       countByHash[hash].Count,
				FirstSeenAt: info.FirstSeenAt,
				LastSeenAt:  info.LastSeenAt,
			})
		}

		return groups, nil
	}

	subQ := r.getHashSubQuery(getHashSubQueryParams{
		table:    "error_groups",
		where:    where,
		duration: req.Duration,
		orderBy:  orderBy(req.Order, true),
		limit:    uint64(req.Limit),
		offset:   uint64(req.Offset),
	})

	infos, err := r.getErrorInfos(ctx, getErrorInfosParams{
		columns: []string{
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		},
		where:    where,
		subQuery: &subQ,
		orderBy:  orderBy(req.Order, false),
	})
	if err != nil {
		return nil, err
	}

	var groups []types.ErrorGroup
	for _, info := range infos {
		groups = append(groups, types.ErrorGroup{
			Hash:        info.Hash,
			Source:      info.Source,
			Message:     info.Message,
			Count:       info.SeenTotal,
			FirstSeenAt: info.FirstSeenAt,
			LastSeenAt:  info.LastSeenAt,
		})
	}

	return groups, nil
}

func (r *repository) GetErrorGroupsTotal(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) (uint64, error) {
	where := sq.Eq{
		"service": req.Service,
	}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}
	if req.Release != nil && *req.Release != "" {
		where["release"] = *req.Release
	}

	return r.getTotal(ctx, getTotalParams{
		where:    where,
		table:    "error_groups",
		duration: req.Duration,
	})
}

func (r *repository) GetNewErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, error) {
	where := sq.Eq{
		"service": req.Service,
	}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	subQ := r.getHashSubQuery(getHashSubQueryParams{
		table:   "error_groups",
		where:   where,
		orderBy: orderBy(req.Order, true),
		limit:   uint64(req.Limit),
		offset:  uint64(req.Offset),
	})
	if req.Release != nil && *req.Release != "" { // new by releases, ignore duration
		subQ = subQ.Having(sq.Eq{
			"any(release)": *req.Release,
			"count()":      1,
		})
	} else if req.Duration != nil && *req.Duration != 0 { // new by duration
		subQ = subQ.Having(sq.GtOrEq{"minMerge(first_seen_at)": r.nowFn().Add(-req.Duration.Abs())})
	}

	infos, err := r.getErrorInfos(ctx, getErrorInfosParams{
		columns: []string{
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		},
		where:    where,
		subQuery: &subQ,
		orderBy:  orderBy(req.Order, false),
	})
	if err != nil {
		return nil, err
	}

	var groups []types.ErrorGroup
	for _, info := range infos {
		groups = append(groups, types.ErrorGroup{
			Hash:        info.Hash,
			Source:      info.Source,
			Message:     info.Message,
			Count:       info.SeenTotal,
			FirstSeenAt: info.FirstSeenAt,
			LastSeenAt:  info.LastSeenAt,
		})
	}

	return groups, nil
}

func (r *repository) GetNewErrorGroupsTotal(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) (uint64, error) {
	subQ := sq.
		Select("_group_hash").
		From("error_groups").
		Where(sq.Eq{"service": req.Service}).
		GroupBy("_group_hash")

	for col, val := range r.queryFilters() {
		subQ = subQ.Where(sq.Eq{col: val})
	}
	if req.Env != nil && *req.Env != "" {
		subQ = subQ.Where(sq.Eq{"env": *req.Env})
	}
	if req.Source != nil && *req.Source != "" {
		subQ = subQ.Where(sq.Eq{"source": *req.Source})
	}

	if req.Release != nil && *req.Release != "" { // new by releases, ignore duration
		subQ = subQ.Having(sq.Eq{
			"any(release)": *req.Release,
			"count()":      1,
		})
	} else if req.Duration != nil && *req.Duration != 0 { // new by duration
		subQ = subQ.Having(sq.GtOrEq{"minMerge(first_seen_at)": r.nowFn().Add(-req.Duration.Abs())})
	}

	q := sq.Select("count()").FromSelect(subQ, "subQ")

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

func (r *repository) GetTopErrorGroups(
	ctx context.Context,
	req types.GetTopErrorGroupsRequest,
) ([]types.TopErrorGroup, error) {
	where := sq.Eq{}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	if req.Duration != nil && *req.Duration != 0 {
		counts, err := r.getErrorCounts(ctx, getErrorCountsParams{
			duration: req.Duration,
			where:    where,
			orderBy:  "count DESC",
			limit:    uint64(req.Limit),
			offset:   uint64(req.Offset),
		})
		if err != nil {
			return nil, err
		}

		if len(counts) == 0 {
			return nil, nil
		}

		where["_group_hash"] = counts.hashes()
		infos, err := r.getErrorInfos(ctx, getErrorInfosParams{
			columns: []string{
				"_group_hash",
				"source",
				"any(message) as message",
			},
			where: where,
		})
		if err != nil {
			return nil, err
		}

		infoByHash := infos.mapByHash()

		var groups []types.TopErrorGroup
		for _, count := range counts {
			info := infoByHash[count.Hash]
			groups = append(groups, types.TopErrorGroup{
				Hash:    count.Hash,
				Source:  info.Source,
				Message: info.Message,
				Count:   count.Count,
			})
		}

		return groups, nil
	}

	subQ := r.getHashSubQuery(getHashSubQueryParams{
		table:   "error_groups_brief",
		where:   where,
		orderBy: orderBy(types.OrderFrequent, true),
		limit:   uint64(req.Limit),
		offset:  uint64(req.Offset),
	})

	infos, err := r.getErrorInfos(ctx, getErrorInfosParams{
		columns: []string{
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
		},
		where:    where,
		subQuery: &subQ,
		orderBy:  orderBy(types.OrderFrequent, false),
	})
	if err != nil {
		return nil, err
	}

	var groups []types.TopErrorGroup
	for _, info := range infos {
		groups = append(groups, types.TopErrorGroup{
			Hash:    info.Hash,
			Source:  info.Source,
			Message: info.Message,
			Count:   info.SeenTotal,
		})
	}

	return groups, nil
}

func (r *repository) GetTopErrorGroupsTotal(
	ctx context.Context,
	req types.GetTopErrorGroupsRequest,
) (uint64, error) {
	where := sq.Eq{}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	return r.getTotal(ctx, getTotalParams{
		table:    "error_groups_brief",
		where:    where,
		duration: req.Duration,
	})
}

func (r *repository) GetErrorHist(
	ctx context.Context,
	req types.GetErrorHistRequest,
) (types.ErrorHist, error) {
	histData := getHistData(req.Duration)

	q := sq.
		Select(
			histData.column,
			"countMerge(counts) as counts",
		).
		From(histData.table).
		GroupBy(histData.column).
		OrderBy(histData.column)

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}
	if req.GroupHash != nil && *req.GroupHash != 0 {
		q = q.Where(sq.Eq{"_group_hash": *req.GroupHash})
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": *req.Source})
	}
	if req.Service != nil && *req.Service != "" {
		q = q.Where(sq.Eq{"service": *req.Service})
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(sq.Eq{"release": *req.Release})
	}
	if req.Duration != nil && *req.Duration != 0 {
		q = q.Where(sq.GtOrEq{histData.column: r.nowFn().Add(-req.Duration.Abs())})
	}

	query, args := q.MustSql()
	metricLabels := []string{"agg_events_10min", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return types.ErrorHist{}, fmt.Errorf("failed to get error hist: %w", err)
	}

	var buckets []types.ErrorHistBucket
	for rows.Next() {
		var bucket types.ErrorHistBucket
		if err := rows.Scan(
			&bucket.Time,
			&bucket.Count,
		); err != nil {
			return types.ErrorHist{}, fmt.Errorf("failed to scan row: %w", err)
		}
		buckets = append(buckets, bucket)
	}

	return types.ErrorHist{
		Buckets:  buckets,
		Interval: histData.interval,
	}, nil
}

func (r *repository) GetErrorDetails(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupDetails, error) {
	q := sq.
		Select(
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
			"max(log_tags) as log_tags",
		).
		From("error_groups").
		Where(sq.Eq{"_group_hash": req.GroupHash}).
		GroupBy("_group_hash", "source")

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": *req.Source})
	}
	if req.Service != nil && *req.Service != "" {
		q = q.Where(sq.Eq{"service": *req.Service})
	}
	if req.Release != nil && *req.Release != "" {
		q = q.Where(sq.Eq{"release": *req.Release})
	}

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	row := r.conn.QueryRow(ctx, metricLabels, query, args...)

	var details types.ErrorGroupDetails
	if err := row.Scan(
		&details.Hash,
		&details.Source,
		&details.Message,
		&details.SeenTotal,
		&details.FirstSeenAt,
		&details.LastSeenAt,
		&details.LogTags,
	); err != nil && !errors.Is(err, sql.ErrNoRows) {
		incErrorMetric(err, metricLabels)
		return details, fmt.Errorf("failed to get error details: %w", err)
	}

	return details, nil
}

type errDetailsCount struct {
	Count   uint64 `ch:"count"`
	Env     string `ch:"env"`
	Source  string `ch:"source"`
	Service string `ch:"service"`
	Release string `ch:"release"`
}

func (r *repository) GetErrorCounts(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupCounts, error) {
	counts := types.ErrorGroupCounts{
		ByEnv:     types.ErrorGroupCount{},
		BySource:  types.ErrorGroupCount{},
		ByService: types.ErrorGroupCount{},
		ByRelease: types.ErrorGroupCount{},
	}

	q := sq.
		Select(
			"countMerge(seen_total) as count",
			"env",
			"source",
			"service",
		).
		From("error_groups").
		Where(sq.Eq{"_group_hash": req.GroupHash}).
		GroupBy("env", "source", "service")

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}

	addFilter := func(col string, val *string) {
		if val != nil && *val != "" {
			q = q.Where(sq.Eq{col: *val})
		}
	}

	addFilter("env", req.Env)
	addFilter("source", req.Source)
	addFilter("service", req.Service)

	// releases only with service
	withRelease := false
	if req.Service != nil && *req.Service != "" {
		withRelease = true
		q = q.Columns("release").GroupBy("release")
		addFilter("release", req.Release)
	}

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return counts, fmt.Errorf("failed to get error counts: %w", err)
	}

	var ec errDetailsCount
	for rows.Next() {
		if err = rows.ScanStruct(&ec); err != nil {
			return counts, fmt.Errorf("failed to scan row: %w", err)
		}
		counts.ByEnv[ec.Env] += ec.Count
		counts.BySource[ec.Source] += ec.Count
		counts.ByService[ec.Service] += ec.Count
		if withRelease {
			counts.ByRelease[ec.Release] += ec.Count
		}
	}

	return counts, nil
}

func (r *repository) GetServices(
	ctx context.Context,
	req types.GetServicesRequest,
) ([]string, error) {
	q := sq.
		Select("service").Distinct().
		From("services").
		Where(sq.NotEq{"service": ""}).
		OrderBy("service")

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}
	if req.Query != "" {
		q = q.Where("startsWith(service, ?)", req.Query)
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
	}
	if req.Limit > 0 {
		q = q.Limit(uint64(req.Limit))
	}
	if req.Offset > 0 {
		q = q.Offset(uint64(req.Offset))
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

func (r *repository) GetReleases(
	ctx context.Context,
	req types.GetReleasesRequest,
) ([]string, error) {
	q := sq.
		Select("release").Distinct().
		From("services").
		Where(sq.And{
			sq.Eq{"service": req.Service},
			sq.NotEq{"release": ""},
		}).
		OrderBy("ttl DESC")

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
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
		releases = append(releases, release)
	}

	return releases, nil
}

func (r *repository) DiffByReleases(
	ctx context.Context,
	req types.DiffByReleasesRequest,
) ([]types.DiffGroup, error) {
	where := sq.Eq{
		"service": req.Service,
		"release": req.Releases,
	}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	groupsQ := sq.
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
		groups    []types.DiffGroup
		idxByHash = map[uint64]int{}
		hashes    []uint64
	)
	for rows.Next() {
		group := types.DiffGroup{
			ReleaseInfos: make(map[string]types.DiffReleaseInfo),
		}
		for _, r := range req.Releases {
			group.ReleaseInfos[r] = types.DiffReleaseInfo{}
		}

		if err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.FirstSeenAt,
			&group.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		groups = append(groups, group)
		hashes = append(hashes, group.Hash)
		idxByHash[group.Hash] = len(groups) - 1
	}

	if len(groups) == 0 {
		return nil, nil
	}

	where["_group_hash"] = hashes
	q := sq.
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
		if err := rows.Scan(
			&hash,
			&release,
			&seenTotal,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		groups[idxByHash[hash]].ReleaseInfos[release] = types.DiffReleaseInfo{
			SeenTotal: seenTotal,
		}
	}

	return groups, nil
}

func (r *repository) DiffByReleasesTotal(
	ctx context.Context,
	req types.DiffByReleasesRequest,
) (uint64, error) {
	where := sq.Eq{
		"service": req.Service,
		"release": req.Releases,
	}
	for col, val := range r.queryFilters() {
		where[col] = val
	}
	if req.Env != nil && *req.Env != "" {
		where["env"] = *req.Env
	}
	if req.Source != nil && *req.Source != "" {
		where["source"] = *req.Source
	}

	return r.getTotal(ctx, getTotalParams{
		table: "error_groups",
		where: where,
	})
}

type getHashSubQueryParams struct {
	table    string
	where    sq.Eq
	duration *time.Duration
	orderBy  string
	limit    uint64
	offset   uint64
}

func (r *repository) getHashSubQuery(params getHashSubQueryParams) sq.SelectBuilder {
	subQ := sq.
		Select("_group_hash").
		From(params.table).
		Where(params.where).
		GroupBy("_group_hash").
		OrderBy(params.orderBy).
		Limit(params.limit).
		Offset(params.offset)

	if params.duration != nil && *params.duration != 0 {
		subQ = subQ.Having(sq.GtOrEq{"maxMerge(last_seen_at)": r.nowFn().Add(-params.duration.Abs())})
	}
	if r.sharded {
		subQ = subQ.Distinct()
	}

	return subQ
}

type getTotalParams struct {
	table    string
	where    sq.Eq
	duration *time.Duration
}

func (r *repository) getTotal(
	ctx context.Context,
	params getTotalParams,
) (uint64, error) {
	q := sq.
		Select("uniq(_group_hash)").
		Where(params.where)

	var table string
	if dur := params.duration; dur != nil && *dur != 0 {
		histData := getHistData(dur)
		q = q.Where(sq.GtOrEq{histData.column: r.nowFn().Add(-dur.Abs())})

		table = histData.table
	} else {
		table = params.table
	}

	query, args := q.From(table).MustSql()
	metricLabels := []string{table, "SELECT"}
	row := r.conn.QueryRow(ctx, metricLabels, query, args...)

	var total uint64
	if err := row.Scan(&total); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		incErrorMetric(err, metricLabels)
		return 0, fmt.Errorf("failed to get total: %w", err)
	}

	return total, nil
}

type errorInfo struct {
	Hash        uint64    `ch:"_group_hash"`
	Source      string    `ch:"source"`
	Message     string    `ch:"message"`
	SeenTotal   uint64    `ch:"seen_total"`
	FirstSeenAt time.Time `ch:"first_seen_at"`
	LastSeenAt  time.Time `ch:"last_seen_at"`
}

type errorInfos []errorInfo

func (i errorInfos) hashes() []uint64 {
	hashes := make([]uint64, 0, len(i))
	for _, v := range i {
		hashes = append(hashes, v.Hash)
	}
	return hashes
}

func (i errorInfos) mapByHash() map[uint64]errorInfo {
	m := make(map[uint64]errorInfo, len(i))
	for _, v := range i {
		m[v.Hash] = v
	}
	return m
}

type getErrorInfosParams struct {
	columns  []string
	where    sq.Eq
	subQuery *sq.SelectBuilder
	orderBy  string
}

func (r *repository) getErrorInfos(
	ctx context.Context,
	params getErrorInfosParams,
) (errorInfos, error) {
	q := sq.
		Select(params.columns...).
		From("error_groups").
		Where(params.where).
		GroupBy("_group_hash", "source")

	if params.subQuery != nil {
		subQ, subArgs := params.subQuery.MustSql()
		q = q.Where(fmt.Sprintf("_group_hash %s (%s)", r.in(), subQ), subArgs...)
	}
	if params.orderBy != "" {
		q = q.OrderBy(params.orderBy)
	}

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups: %w", err)
	}

	var infos errorInfos
	for rows.Next() {
		var info errorInfo
		if err = rows.ScanStruct(&info); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}

type errorCount struct {
	Hash  uint64 `ch:"_group_hash"`
	Count uint64 `ch:"count"`
}

type errorCounts []errorCount

func (c errorCounts) hashes() []uint64 {
	hashes := make([]uint64, 0, len(c))
	for _, v := range c {
		hashes = append(hashes, v.Hash)
	}
	return hashes
}

func (c errorCounts) mapByHash() map[uint64]errorCount {
	m := make(map[uint64]errorCount, len(c))
	for _, v := range c {
		m[v.Hash] = v
	}
	return m
}

type getErrorCountsParams struct {
	duration *time.Duration
	where    sq.Eq
	orderBy  string
	limit    uint64
	offset   uint64
}

func (r *repository) getErrorCounts(
	ctx context.Context,
	params getErrorCountsParams,
) (errorCounts, error) {
	histData := getHistData(params.duration)

	q := sq.
		Select(
			"_group_hash",
			"countMerge(counts) as count",
		).
		From(histData.table).
		Where(params.where).
		Where(sq.GtOrEq{histData.column: r.nowFn().Add(-params.duration.Abs())}).
		GroupBy("_group_hash")

	if params.orderBy != "" {
		q = q.OrderBy(params.orderBy)
	}
	if params.limit > 0 {
		q = q.Limit(params.limit)
	}
	if params.offset > 0 {
		q = q.Offset(params.offset)
	}

	query, args := q.MustSql()
	metricLabels := []string{histData.table, "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups count: %w", err)
	}

	var counts errorCounts
	for rows.Next() {
		var ec errorCount
		if err = rows.ScanStruct(&ec); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		counts = append(counts, ec)
	}

	return counts, nil
}

func (r *repository) in() string {
	if r.sharded {
		return "GLOBAL IN"
	}
	return "IN"
}

func orderBy(o types.ErrorGroupsOrder, sub bool) string {
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
		return seenTotal
	case types.OrderLatest:
		return lastSeenAt
	case types.OrderOldest:
		return firstSeenAt
	}
	return seenTotal
}

type histData struct {
	table    string
	column   string
	interval uint64
}

func getHistData(duration *time.Duration) histData {
	const (
		table_10min = "agg_events_10min"
		table_1d    = "agg_events_1d"

		startDate    = "start_date"
		startOfHour  = "toStartOfHour(start_date)"
		startOfDay   = "toStartOfDay(start_date)"
		startOfWeek  = "toStartOfWeek(start_date)"
		startOfMonth = "toStartOfMonth(start_date)"

		_10min = 10 * time.Minute
		hour   = time.Hour
		day    = 24 * hour
		week   = 7 * day
		month  = 31 * day
	)

	data := func(table, column string, interval time.Duration) histData {
		return histData{table: table, column: column, interval: uint64(interval.Seconds())}
	}

	if duration == nil || *duration == 0 {
		return data(table_1d, startOfMonth, month)
	}

	// try get ~30 buckets
	d := *duration
	switch {
	case d <= 5*time.Hour:
		return data(table_10min, startDate, _10min)
	case d <= day:
		return data(table_10min, startOfHour, hour)
	case d <= month:
		return data(table_10min, startOfDay, day)
	case d <= 7*month:
		return data(table_1d, startOfWeek, week)
	default:
		return data(table_1d, startOfMonth, month)
	}
}
