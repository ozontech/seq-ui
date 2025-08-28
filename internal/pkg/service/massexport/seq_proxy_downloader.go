package massexport

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type seqProxyDownloader struct {
	cfg    config.SeqProxyDownloader
	client seqdb.Client
}

const defaultDelay = 1 * time.Second

func newSeqProxyDownloader(client seqdb.Client, cfg config.SeqProxyDownloader) *seqProxyDownloader {
	if cfg.Delay <= 0 {
		cfg.Delay = defaultDelay
	}

	cfg.InitialRetryBackoff = max(cfg.InitialRetryBackoff, cfg.Delay)
	cfg.MaxRetryBackoff = max(cfg.MaxRetryBackoff, cfg.InitialRetryBackoff)

	return &seqProxyDownloader{
		cfg:    cfg,
		client: client,
	}
}

func (d *seqProxyDownloader) Search(ctx context.Context, sessionID string, req *seqapi.SearchRequest) (*seqapi.SearchResponse, error) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	bo.Multiplier = 2
	bo.InitialInterval = d.cfg.InitialRetryBackoff
	bo.MaxInterval = d.cfg.MaxRetryBackoff
	bo.Reset()

	for {
		start := time.Now()
		resp, err := d.client.Search(ctx, req)
		if err == nil {
			duration := time.Since(start)

			logger.Info(
				"got search response for mass export",
				zap.Duration("duration", duration),
				zap.String("session_id", sessionID),
			)
			metric.MassExportSeqDBSearchDuration.WithLabelValues(sessionID).Observe(duration.Seconds())

			sleepCtx(ctx, d.cfg.Delay-duration)

			return resp, nil
		}

		if status.Code(err) != codes.ResourceExhausted {
			return nil, err
		}
		logger.Warn(
			"rate limit happened while mass export",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)

		sleepCtx(ctx, bo.NextBackOff())
	}
}

func sleepCtx(ctx context.Context, duration time.Duration) {
	if duration == 0 {
		return
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
