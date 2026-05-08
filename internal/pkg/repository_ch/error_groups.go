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
		aggTable, startDate := getHistData(req.Duration)

		if req.Order == types.OrderFrequent {
			countQ := sq.
				Select(
					"_group_hash",
					"countMerge(counts) as count",
				).
				From(aggTable).
				Where(where).
				Where(sq.GtOrEq{startDate: r.nowFn().Add(-req.Duration.Abs())}).
				GroupBy("_group_hash").
				OrderBy("count DESC").
				Limit(uint64(req.Limit)).
				Offset(uint64(req.Offset))

			countQuery, countArgs := countQ.MustSql()
			metricLabels := []string{aggTable, "SELECT"}
			countRows, err := r.conn.Query(ctx, metricLabels, countQuery, countArgs...)
			if err != nil {
				incErrorMetric(err, metricLabels)
				return nil, fmt.Errorf("failed to get error groups count: %w", err)
			}

			var (
				groups    []types.ErrorGroup
				idxByHash = map[uint64]int{}
				hashes    []uint64
			)
			for countRows.Next() {
				var group types.ErrorGroup
				if err = countRows.Scan(
					&group.Hash,
					&group.Count,
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
					"source",
					"any(message) as message",
					"minMerge(first_seen_at) as first_seen_at",
					"maxMerge(last_seen_at) as last_seen_at",
				).
				From("error_groups").
				Where(where).
				GroupBy("_group_hash", "source")

			query, args := q.MustSql()
			metricLabels = []string{"error_groups", "SELECT"}
			rows, err := r.conn.Query(ctx, metricLabels, query, args...)
			if err != nil {
				incErrorMetric(err, metricLabels)
				return nil, fmt.Errorf("failed to get error groups: %w", err)
			}

			var (
				hash                uint64
				message, source     string
				firstSeen, lastSeen time.Time
			)
			for rows.Next() {
				if err = rows.Scan(
					&hash,
					&source,
					&message,
					&firstSeen,
					&lastSeen,
				); err != nil {
					return nil, fmt.Errorf("failed to scan row: %w", err)
				}

				idx := idxByHash[hash]
				groups[idx].Source = source
				groups[idx].Message = message
				groups[idx].FirstSeenAt = firstSeen
				groups[idx].LastSeenAt = lastSeen
			}

			return groups, nil
		}

		subQ := sq.
			Select("_group_hash").
			From("error_groups").
			Where(where).
			Having(sq.GtOrEq{"maxMerge(last_seen_at)": r.nowFn().Add(-req.Duration.Abs())}).
			GroupBy("_group_hash").
			OrderBy(orderBy(req.Order, true)).
			Limit(uint64(req.Limit)).
			Offset(uint64(req.Offset))

		in := "IN"
		if r.sharded {
			subQ = subQ.Distinct()
			in = "GLOBAL IN"
		}

		subQuery, subArgs := subQ.MustSql()

		q := sq.
			Select(
				"_group_hash",
				"source",
				"any(message) as message",
				"minMerge(first_seen_at) as first_seen_at",
				"maxMerge(last_seen_at) as last_seen_at",
			).
			From("error_groups").
			Where(where).
			Where(fmt.Sprintf("_group_hash %s (%s)", in, subQuery), subArgs...).
			GroupBy("_group_hash", "source").
			OrderBy(orderBy(req.Order, false))

		query, args := q.MustSql()
		metricLabels := []string{"error_groups", "SELECT"}
		rows, err := r.conn.Query(ctx, metricLabels, query, args...)
		if err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to get error groups: %w", err)
		}

		var (
			groups    []types.ErrorGroup
			idxByHash = map[uint64]int{}
			hashes    []uint64
		)
		for rows.Next() {
			var group types.ErrorGroup
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
		countQ := sq.
			Select(
				"_group_hash",
				"countMerge(counts) as count",
			).
			From(aggTable).
			Where(where).
			GroupBy("_group_hash")

		countQuery, countArgs := countQ.MustSql()
		metricLabels = []string{aggTable, "SELECT"}
		countRows, err := r.conn.Query(ctx, metricLabels, countQuery, countArgs...)
		if err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to get error groups count: %w", err)
		}

		var hash, count uint64
		for countRows.Next() {
			if err = countRows.Scan(
				&hash,
				&count,
			); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}

			groups[idxByHash[hash]].Count = count
		}

		return groups, nil
	}

	subQ := sq.
		Select("_group_hash").
		From("error_groups").
		Where(where).
		GroupBy("_group_hash").
		OrderBy(orderBy(req.Order, true)).
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	in := "IN"
	if r.sharded {
		subQ = subQ.Distinct()
		in = "GLOBAL IN"
	}

	subQuery, subArgs := subQ.MustSql()

	q := sq.
		Select(
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		).
		From("error_groups").
		Where(where).
		Where(fmt.Sprintf("_group_hash %s (%s)", in, subQuery), subArgs...).
		GroupBy("_group_hash", "source").
		OrderBy(orderBy(req.Order, false))

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups: %w", err)
	}

	var groups []types.ErrorGroup
	for rows.Next() {
		var group types.ErrorGroup
		if err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.Count,
			&group.FirstSeenAt,
			&group.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (r *repository) GetErrorGroupsTotal(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) (uint64, error) {
	subQ := sq.
		Select("maxMerge(last_seen_at) AS last_seen_at").
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
	if req.Release != nil && *req.Release != "" {
		subQ = subQ.Where(sq.Eq{"release": *req.Release})
	}
	if req.Duration != nil && *req.Duration != 0 {
		subQ = subQ.Having(sq.GtOrEq{"last_seen_at": r.nowFn().Add(-req.Duration.Abs())})
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

	subQ := sq.
		Select("_group_hash").
		From("error_groups").
		Where(where).
		GroupBy("_group_hash").
		OrderBy(orderBy(req.Order, true)).
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	in := "IN"
	if r.sharded {
		subQ = subQ.Distinct()
		in = "GLOBAL IN"
	}

	if req.Release != nil && *req.Release != "" { // new by releases, ignore duration
		subQ = subQ.Having(sq.Eq{
			"any(release)": *req.Release,
			"count()":      1,
		})
	} else if req.Duration != nil && *req.Duration != 0 { // new by duration
		subQ = subQ.Having(sq.GtOrEq{"minMerge(first_seen_at)": r.nowFn().Add(-req.Duration.Abs())})
	}

	subQuery, subArgs := subQ.MustSql()

	q := sq.
		Select(
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
			"minMerge(first_seen_at) as first_seen_at",
			"maxMerge(last_seen_at) as last_seen_at",
		).
		From("error_groups").
		Where(where).
		Where(fmt.Sprintf("_group_hash %s (%s)", in, subQuery), subArgs...).
		GroupBy("_group_hash", "source").
		OrderBy(orderBy(req.Order, false))

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get new error groups: %w", err)
	}

	var groups []types.ErrorGroup
	for rows.Next() {
		var group types.ErrorGroup
		if err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.Count,
			&group.FirstSeenAt,
			&group.LastSeenAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		groups = append(groups, group)
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
		return 0, fmt.Errorf("failed to get new error groups count: %w", err)
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
		aggTable, startDate := getHistData(req.Duration)

		countQ := sq.
			Select(
				"_group_hash",
				"countMerge(counts) as count",
			).
			From(aggTable).
			Where(where).
			Where(sq.GtOrEq{startDate: r.nowFn().Add(-req.Duration.Abs())}).
			GroupBy("_group_hash").
			OrderBy("count DESC").
			Limit(uint64(req.Limit)).
			Offset(uint64(req.Offset))

		countQuery, countArgs := countQ.MustSql()
		metricLabels := []string{aggTable, "SELECT"}
		countRows, err := r.conn.Query(ctx, metricLabels, countQuery, countArgs...)
		if err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to get error groups count: %w", err)
		}

		var (
			groups    []types.TopErrorGroup
			idxByHash = map[uint64]int{}
			hashes    []uint64
		)
		for countRows.Next() {
			var group types.TopErrorGroup
			if err = countRows.Scan(
				&group.Hash,
				&group.Count,
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
				"source",
				"any(message) as message",
			).
			From("error_groups").
			Where(where).
			GroupBy("_group_hash", "source")

		query, args := q.MustSql()
		metricLabels = []string{"error_groups", "SELECT"}
		rows, err := r.conn.Query(ctx, metricLabels, query, args...)
		if err != nil {
			incErrorMetric(err, metricLabels)
			return nil, fmt.Errorf("failed to get error groups: %w", err)
		}

		var (
			hash            uint64
			message, source string
		)
		for rows.Next() {
			if err = rows.Scan(
				&hash,
				&source,
				&message,
			); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}

			idx := idxByHash[hash]
			groups[idx].Message = message
			groups[idx].Source = source
		}

		return groups, nil
	}

	subQ := sq.
		Select("_group_hash").
		From("error_groups_brief").
		Where(where).
		GroupBy("_group_hash").
		OrderBy(orderBy(types.OrderFrequent, true)).
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	in := "IN"
	if r.sharded {
		subQ = subQ.Distinct()
		in = "GLOBAL IN"
	}

	subQuery, subArgs := subQ.MustSql()

	q := sq.
		Select(
			"_group_hash",
			"source",
			"any(message) as message",
			"countMerge(seen_total) as seen_total",
		).
		From("error_groups").
		Where(where).
		Where(fmt.Sprintf("_group_hash %s (%s)", in, subQuery), subArgs...).
		GroupBy("_group_hash", "source").
		OrderBy(orderBy(types.OrderFrequent, false))

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return nil, fmt.Errorf("failed to get error groups: %w", err)
	}

	var groups []types.TopErrorGroup
	for rows.Next() {
		var group types.TopErrorGroup
		if err = rows.Scan(
			&group.Hash,
			&group.Source,
			&group.Message,
			&group.Count,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (r *repository) GetTopErrorGroupsTotal(
	ctx context.Context,
	req types.GetTopErrorGroupsRequest,
) (uint64, error) {
	q := sq.Select("uniq(_group_hash)")

	for col, val := range r.queryFilters() {
		q = q.Where(sq.Eq{col: val})
	}
	if req.Env != nil && *req.Env != "" {
		q = q.Where(sq.Eq{"env": *req.Env})
	}
	if req.Source != nil && *req.Source != "" {
		q = q.Where(sq.Eq{"source": *req.Source})
	}

	var table string
	if req.Duration != nil && *req.Duration != 0 {
		aggTable, startDate := getHistData(req.Duration)
		q = q.Where(sq.GtOrEq{startDate: r.nowFn().Add(-req.Duration.Abs())})

		table = aggTable
	} else {
		table = "error_groups_brief"
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
		return 0, fmt.Errorf("failed to get top error groups count: %w", err)
	}

	return total, nil
}

func (r *repository) GetErrorHist(
	ctx context.Context,
	req types.GetErrorHistRequest,
) ([]types.ErrorHistBucket, error) {
	table, startDate := getHistData(req.Duration)

	q := sq.
		Select(
			startDate,
			"countMerge(counts) as counts",
		).
		From(table).
		GroupBy(startDate).
		OrderBy(startDate)

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
		q = q.Where(sq.GtOrEq{startDate: r.nowFn().Add(-req.Duration.Abs())})
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
		if err := rows.Scan(
			&bucket.Time,
			&bucket.Count,
		); err != nil {
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

type errCounts struct {
	count   uint64
	env     string
	source  string
	service string
	release string
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
		addFilter("release", req.Release)
		q = q.Columns("release").GroupBy("release")
	}

	query, args := q.MustSql()
	metricLabels := []string{"error_groups", "SELECT"}
	rows, err := r.conn.Query(ctx, metricLabels, query, args...)
	if err != nil {
		incErrorMetric(err, metricLabels)
		return counts, fmt.Errorf("failed to get error counts: %w", err)
	}

	var ec errCounts
	for rows.Next() {
		if err = rows.ScanStruct(&ec); err != nil {
			return counts, fmt.Errorf("failed to scan row: %w", err)
		}
		counts.ByEnv[ec.env] += ec.count
		counts.BySource[ec.source] += ec.count
		counts.ByService[ec.service] += ec.count
		if withRelease {
			counts.ByRelease[ec.release] += ec.count
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
	subQ := sq.
		Select("_group_hash").
		From("error_groups").
		Where(sq.Eq{
			"service": req.Service,
			"release": req.Releases,
		}).
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

func getHistData(duration *time.Duration) (string, string) {
	const (
		table_10min = "agg_events_10min"
		table_1d    = "agg_events_1d"

		startDate    = "start_date"
		startOfHour  = "toStartOfHour(start_date)"
		startOfDay   = "toStartOfDay(start_date)"
		startOfWeek  = "toStartOfWeek(start_date)"
		startOfMonth = "toStartOfMonth(start_date)"

		day   = 24 * time.Hour
		month = 30 * day
	)

	if duration == nil || *duration == 0 {
		return table_1d, startOfMonth
	}

	// try get ~30 buckets
	d := *duration
	switch {
	case d <= 5*time.Hour:
		return table_10min, startDate
	case d <= day:
		return table_10min, startOfHour
	case d <= month:
		return table_10min, startOfDay
	case d <= 7*month:
		return table_1d, startOfWeek
	default:
		return table_1d, startOfMonth
	}
}
