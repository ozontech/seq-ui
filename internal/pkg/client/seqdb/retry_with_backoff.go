package seqdb

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type retriableReqFn func() (any, bool, error)

type tryWithBackoffParams struct {
	maxRetries          int
	initialRetryBackoff time.Duration
	maxRetryBackoff     time.Duration
}

func trySendRequestWithBackoff(ctx context.Context, fn retriableReqFn, params tryWithBackoffParams) (any, bool, error) {
	var err error
	var resp any
	var retry bool

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	bo.Multiplier = 2
	bo.InitialInterval = params.initialRetryBackoff
	bo.MaxInterval = params.maxRetryBackoff
	bo.Reset()
	// +1 is for the very first attempt before retries
	for i := 0; i < params.maxRetries+1; i++ {
		resp, retry, err = fn()
		if !retry {
			return resp, true, err
		}

		// no need to call next backoff if the initial interval is 0
		// no need to sleep after the last attempt
		if bo.InitialInterval != 0 && i != params.maxRetries {
			sleepCtx(ctx, bo.NextBackOff())
		}
	}
	return resp, false, fmt.Errorf("send request: %w", err)
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
