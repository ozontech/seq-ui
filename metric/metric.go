package metric

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	seqUINS           = "seq_ui_server"
	serverSubsys      = "server"
	seqDBClientSubsys = "seq_db_client"
	authSubsys        = "auth"
	repoSubsys        = "repository"
	clickHouseSubsys  = "clickhouse"
	massExportSubsys  = "mass_export"

	componentLabel  = "component"
	methodLabel     = "method"
	statusCodeLabel = "status_code"
	errorTypeLabel  = "error_type"
	opLabel         = "op"
	tableLabel      = "table"
	queryLabel      = "query"
	sessionIDLabel  = "session_id"
)

var (
	defaultBuckets = prometheus.ExponentialBuckets(0.002, 2, 16)

	// server metrics
	ServerRequestReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "requests_received_total",
		Help:      "",
	}, []string{componentLabel, methodLabel})
	ServerRequestHandled = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "requests_handled_total",
		Help:      "",
	}, []string{componentLabel, methodLabel, statusCodeLabel})
	ServerRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "requests_duration_seconds",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{componentLabel, methodLabel, statusCodeLabel})
	ServerRequestPanics = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "requests_panics_total",
		Help:      "",
	})
	ServerRateLimits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "requests_rate_limits_total",
		Help:      "",
	})
	ServerExportRequestLimits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "export_requests_limits_total",
		Help:      "",
	})
	ServerCacheInmemoryHits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "cache_inmemory_hits_total",
		Help:      "",
	})
	ServerCacheInmemoryMisses = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "cache_inmemory_misses_total",
		Help:      "",
	})
	ServerCacheRedisHits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "cache_redis_hits_total",
		Help:      "",
	})
	ServerCacheRedisMisses = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "cache_redis_misses_total",
		Help:      "",
	})
	// client metrics
	SeqDBClientRequestSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "requests_sent_total",
		Help:      "",
	}, []string{methodLabel})
	SeqDBClientResponseReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "responses_received_total",
		Help:      "",
	}, []string{methodLabel, statusCodeLabel})
	SeqDBClientRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "requests_sent_duration_seconds",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{methodLabel, statusCodeLabel})
	SeqDBClientRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "requests_errors_total",
		Help:      "",
	}, []string{methodLabel, errorTypeLabel})
	SeqDBClientPartialResponse = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "partial_responses_total",
		Help:      "",
	}, []string{methodLabel, errorTypeLabel})
	SeqDBClientStreamError = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "stream_errors_total",
		Help:      "",
	}, []string{methodLabel, opLabel})
	SeqDBClientEmptyDataResponse = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: seqDBClientSubsys,
		Name:      "empty_data_responses_total",
		Help:      "",
	})
	// auth metrics
	AuthVerifyDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: authSubsys,
		Name:      "verify_duration_seconds",
		Help:      "",
		Buckets:   defaultBuckets,
	})
	// repository metrics
	RepositoryRequestSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: repoSubsys,
		Name:      "requests_sent_total",
		Help:      "",
	}, []string{tableLabel, queryLabel})
	RepositoryRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: repoSubsys,
		Name:      "requests_sent_duration_seconds",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{tableLabel, queryLabel})
	RepositoryRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: repoSubsys,
		Name:      "requests_errors_total",
		Help:      "",
	}, []string{tableLabel, queryLabel, errorTypeLabel})
	UnauthorizedClientsRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: serverSubsys,
		Name:      "unauthorized_clients_requests_count",
		Help:      "",
	}, []string{"client"})

	// clickhouse repository metrics
	ClickHouseRequestSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: clickHouseSubsys,
		Name:      "requests_sent_total",
		Help:      "",
	}, []string{tableLabel, queryLabel})
	ClickHouseRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: clickHouseSubsys,
		Name:      "requests_sent_duration_seconds",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{tableLabel, queryLabel})
	ClickHouseRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: seqUINS,
		Subsystem: clickHouseSubsys,
		Name:      "requests_errors_total",
		Help:      "",
	}, []string{tableLabel, queryLabel, errorTypeLabel})

	// mass export metrics
	MassExportSeqDBSearchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: massExportSubsys,
		Name:      "seq_db_search_duration",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{sessionIDLabel})

	MassExportOnePartExportDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: seqUINS,
		Subsystem: massExportSubsys,
		Name:      "one_part_export_duration",
		Help:      "",
		Buckets:   defaultBuckets,
	}, []string{sessionIDLabel})
)

// HandledIncomingRequest handles metrics for processed incoming request.
func HandledIncomingRequest(ctx context.Context, component, method, statusCode string, took time.Duration) {
	ctxErr := ctx.Err()
	if errors.Is(ctxErr, context.Canceled) {
		statusCode = context.Canceled.Error()
	} else if errors.Is(ctxErr, context.DeadlineExceeded) {
		statusCode = context.DeadlineExceeded.Error()
	}
	ServerRequestDuration.WithLabelValues(component, method, statusCode).Observe(took.Seconds())
	ServerRequestHandled.WithLabelValues(component, method, statusCode).Inc()
}
