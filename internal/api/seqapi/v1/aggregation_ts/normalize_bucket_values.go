package aggregation_ts

import (
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func NormalizeBucketValues(aggregations []*seqapi.Aggregation, aggIntervals, bucketUnits []time.Duration) {
	for i, agg := range aggregations {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == 0 {
			continue
		}

		bucketUnitDenominator := time.Second
		if bucketUnits[i] != 0 {
			bucketUnitDenominator = bucketUnits[i]
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}

			*bucket.Value = *bucket.Value * float64(bucketUnitDenominator) / float64(aggIntervals[i])
		}
	}
}
