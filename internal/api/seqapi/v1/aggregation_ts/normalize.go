package aggregation_ts

import (
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type aggQuery interface {
	GetFunc() seqapi.AggFunc
	GetInterval() string
	GetBucketUnit() string
}

func NormalizeBuckets[T aggQuery](aggQueries []T, aggs []*seqapi.Aggregation, defaultBucketUnit time.Duration) error {
	bucketUnits, err := getBucketUnits(aggQueries, defaultBucketUnit)
	if err != nil {
		return fmt.Errorf("failed to get bucket units: %w", err)
	}

	aggIntervals, err := getIntervals(aggQueries)
	if err != nil {
		return fmt.Errorf("failed to get intervals: %w", err)
	}

	for i, agg := range aggs {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == 0 {
			continue
		}

		bucketUnitDenominator := time.Second
		if bucketUnits[i] != 0 {
			bucketUnitDenominator = bucketUnits[i]
			agg.BucketUnit = bucketUnits[i].String()
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}

			*bucket.Value = *bucket.Value * float64(bucketUnitDenominator) / float64(aggIntervals[i])
		}
	}

	return nil
}

func getBucketUnits[T aggQuery](aggQueries []T, defaultBucketUnit time.Duration) ([]time.Duration, error) {
	aggBucketUnits := make([]time.Duration, 0, len(aggQueries))
	for _, agg := range aggQueries {
		if agg.GetFunc() != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggBucketUnits = append(aggBucketUnits, 0)
			continue
		}
		bucketUnitRaw := agg.GetBucketUnit()
		if bucketUnitRaw == "" {
			aggBucketUnits = append(aggBucketUnits, defaultBucketUnit)
			continue
		}

		bucketUnit, err := time.ParseDuration(bucketUnitRaw)
		if err != nil {
			return nil, err
		}

		aggBucketUnits = append(aggBucketUnits, bucketUnit)
	}

	return aggBucketUnits, nil
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
