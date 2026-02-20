package aggregation_ts

import (
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func GetBucketUnits(aggregations []*seqapi.AggregationQuery, defaultBucketUnit time.Duration) ([]time.Duration, error) {
	aggBucketUnits := make([]time.Duration, 0, len(aggregations))
	for _, agg := range aggregations {
		if agg.Func != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggBucketUnits = append(aggBucketUnits, 0)
			continue
		}
		if agg.BucketUnit == nil {
			aggBucketUnits = append(aggBucketUnits, defaultBucketUnit)
			continue
		}

		bucketUnit, err := time.ParseDuration(*agg.BucketUnit)
		if err != nil {
			return nil, err
		}

		aggBucketUnits = append(aggBucketUnits, bucketUnit)
	}

	return aggBucketUnits, nil
}
