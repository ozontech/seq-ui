package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/seq-ui/metric"
)

type pool struct {
	pool           *pgxpool.Pool
	requestTimeout time.Duration
}

func newPool(p *pgxpool.Pool, requestTimeout time.Duration) *pool {
	return &pool{
		pool:           p,
		requestTimeout: requestTimeout,
	}
}

func (p *pool) query(ctx context.Context, metricLabels []string, query string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, p.requestTimeout)
	defer cancel()

	metric.RepositoryRequestSent.WithLabelValues(metricLabels...).Inc()
	start := time.Now()
	rows, err := p.pool.Query(ctx, query, args...)
	took := time.Since(start)
	metric.RepositoryRequestDuration.WithLabelValues(metricLabels...).Observe(took.Seconds())

	return rows, err
}

func (p *pool) queryRow(ctx context.Context, metricLabels []string, query string, args ...any) pgx.Row {
	ctx, cancel := context.WithTimeout(ctx, p.requestTimeout)
	defer cancel()

	metric.RepositoryRequestSent.WithLabelValues(metricLabels...).Inc()
	start := time.Now()
	row := p.pool.QueryRow(ctx, query, args...)
	took := time.Since(start)
	metric.RepositoryRequestDuration.WithLabelValues(metricLabels...).Observe(took.Seconds())

	return row
}

func (p *pool) exec(ctx context.Context, metricLabels []string, query string, args ...any) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(ctx, p.requestTimeout)
	defer cancel()

	metric.RepositoryRequestSent.WithLabelValues(metricLabels...).Inc()
	start := time.Now()
	tag, err := p.pool.Exec(ctx, query, args...)
	took := time.Since(start)
	metric.RepositoryRequestDuration.WithLabelValues(metricLabels...).Observe(took.Seconds())

	return tag, err
}
