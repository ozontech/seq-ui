package aggregation_ts

import (
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type aggQuery interface {
	GetFunc() seqapi.AggFunc
	GetInterval() string
	GetTargetBucketRate() string
}

func NormalizeBuckets[T aggQuery](aggQueries []T, aggs []*seqapi.Aggregation) error {
	targetBucketRates, err := getTargetBucketRates(aggQueries)
	if err != nil {
		return fmt.Errorf("failed to get bucket units: %w", err)
	}

	aggIntervals, err := getIntervals(aggQueries)
	if err != nil {
		return fmt.Errorf("failed to get intervals: %w", err)
	}

	for i, agg := range aggs {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == 0 || targetBucketRates[i] == 0 {
			continue
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}

			*bucket.Value = *bucket.Value * float64(targetBucketRates[i]) / float64(aggIntervals[i])
		}
		agg.TargetBucketRate = targetBucketRates[i].String()
	}

	return nil
}

func getTargetBucketRates[T aggQuery](aggQueries []T) ([]time.Duration, error) {
	aggTargetBucketRates := make([]time.Duration, 0, len(aggQueries))
	for _, agg := range aggQueries {
		targetBucketRateRaw := agg.GetTargetBucketRate()
		if agg.GetFunc() != seqapi.AggFunc_AGG_FUNC_COUNT || targetBucketRateRaw == "" {
			aggTargetBucketRates = append(aggTargetBucketRates, 0)
			continue
		}

		targetBucketRate, err := time.ParseDuration(targetBucketRateRaw)
		if err != nil {
			return nil, err
		}

		aggTargetBucketRates = append(aggTargetBucketRates, targetBucketRate)
	}

	return aggTargetBucketRates, nil
}

func getIntervals[T aggQuery](aggQueries []T) ([]time.Duration, error) {
	aggIntervals := make([]time.Duration, 0, len(aggQueries))
	for _, agg := range aggQueries {
		if agg.GetFunc() != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggIntervals = append(aggIntervals, 0)
			continue
		}
		intervalRaw := agg.GetInterval()
		if intervalRaw == "" {
			aggIntervals = append(aggIntervals, time.Second)
			continue
		}

		interval, err := time.ParseDuration(intervalRaw)
		if err != nil {
			return nil, err
		}

		aggIntervals = append(aggIntervals, interval)
	}

	return aggIntervals, nil
}
