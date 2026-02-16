package aggregationts

import (
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func NormalizeBucketValues(aggregations []*seqapi.Aggregation, aggIntervals, bucketUnits []*string) error {
	for i, agg := range aggregations {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == nil {
			continue
		}

		interval, err := time.ParseDuration(*aggIntervals[i])
		if err != nil {
			return fmt.Errorf("failed to parse aggregation interval: %w", err)
		}

		BucketUnit := time.Second
		if i < len(bucketUnits) && bucketUnits[i] != nil {
			BucketUnit, err = time.ParseDuration(*bucketUnits[i])
			if err != nil {
				return fmt.Errorf("failed to parse bucket quantity: %w", err)
			}
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}
			*bucket.Value = *bucket.Value * float64(BucketUnit) / float64(interval)
		}
	}

	return nil
}
