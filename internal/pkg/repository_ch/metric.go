package repositorych

import (
	"context"
	"errors"
	"os"

	"github.com/ozontech/seq-ui/metric"
)

const (
	queryErrorOther       = "other"
	queryErrorTimeout     = "timeout"
	queryErrorCtxCanceled = "context canceled"
)

func incErrorMetric(err error, metricLabels []string) {
	if err == nil {
		return
	}

	var queryErr string
	switch {
	case errors.Is(err, context.Canceled):
		queryErr = queryErrorCtxCanceled
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, os.ErrDeadlineExceeded):
		queryErr = queryErrorTimeout
	default:
		queryErr = queryErrorOther
	}

	metricLabels = append(metricLabels, queryErr)
	metric.ClickHouseRequestError.WithLabelValues(metricLabels...).Inc()
}
