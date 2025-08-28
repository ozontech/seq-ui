package repository

import (
	"context"
	"errors"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/metric"
)

const (
	queryErrorOther            = "other"
	queryErrorTimeout          = "timeout"
	queryErrorCtxCanceled      = "context canceled"
	queryErrorNotFound         = "not found"
	queryErrorPermissionDenied = "permission denied"
)

func incErrorMetric(err error, metricLabels []string) {
	if err == nil {
		return
	}

	var queryErr string
	switch {
	case errors.Is(err, context.Canceled):
		queryErr = queryErrorCtxCanceled
	case errors.Is(err, context.DeadlineExceeded):
		queryErr = queryErrorTimeout
	case errors.Is(err, types.ErrNotFound):
		queryErr = queryErrorNotFound
	case errors.Is(err, types.ErrPermissionDenied):
		queryErr = queryErrorPermissionDenied
	default:
		queryErr = queryErrorOther
	}

	metricLabels = append(metricLabels, queryErr)
	metric.RepositoryRequestError.WithLabelValues(metricLabels...).Inc()
}
