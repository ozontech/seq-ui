package repositorych

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/ozontech/seq-ui/metric"
)

type conn struct {
	conn driver.Conn
}

func newConn(c driver.Conn) *conn {
	return &conn{
		conn: c,
	}
}

func (c *conn) Query(ctx context.Context, metricLabels []string, query string, args ...any) (driver.Rows, error) {
	metric.ClickHouseRequestSent.WithLabelValues(metricLabels...).Inc()
	start := time.Now()
	rows, err := c.conn.Query(ctx, query, args...)
	took := time.Since(start)
	metric.ClickHouseRequestDuration.WithLabelValues(metricLabels...).Observe(took.Seconds())

	return rows, err
}

func (c *conn) QueryRow(ctx context.Context, metricLabels []string, query string, args ...any) driver.Row {
	metric.ClickHouseRequestSent.WithLabelValues(metricLabels...).Inc()
	start := time.Now()
	row := c.conn.QueryRow(ctx, query, args...)
	took := time.Since(start)
	metric.ClickHouseRequestDuration.WithLabelValues(metricLabels...).Observe(took.Seconds())

	return row
}
