package aggregationts

import (
	"fmt"
	"strings"
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

const bucketUnitPrefix = "count/"

func NormalizeBucketValues(aggregations []*seqapi.Aggregation, aggIntervals, bucketUnits []*string) error {
	for i, agg := range aggregations {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == nil {
			continue
		}

		interval, err := time.ParseDuration(*aggIntervals[i])
		if err != nil {
			return fmt.Errorf("failed to parse aggregation interval: %w", err)
		}

		BucketUnitDenominator := time.Second
		if i < len(bucketUnits) && bucketUnits[i] != nil {
			agg.BucketUnit = *bucketUnits[i]

			BucketUnitDenominator, err = parseBucketUnitDenominator(bucketUnits[i])
			if err != nil {
				return fmt.Errorf("failed to parse bucket unit: %w", err)
			}
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}
			*bucket.Value = *bucket.Value * float64(BucketUnitDenominator) / float64(interval)
		}
	}

	return nil
}

func parseBucketUnitDenominator(bucketUnit *string) (time.Duration, error) {
	bucketUnitDenominator, err := time.ParseDuration(strings.TrimPrefix(*bucketUnit, bucketUnitPrefix))
	if err != nil {
		return 0, err
	}

	return bucketUnitDenominator, nil
}
