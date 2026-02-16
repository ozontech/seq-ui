package aggregationts

import "github.com/ozontech/seq-ui/pkg/seqapi/v1"

func GetBucketUnits(aggregations []*seqapi.AggregationQuery) []*string {
	aggBucketUnits := make([]*string, 0, len(aggregations))
	for _, agg := range aggregations {
		if agg.Func != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggBucketUnits = append(aggBucketUnits, nil)
			continue
		}
		aggBucketUnits = append(aggBucketUnits, agg.BucketUnit)
	}

	return aggBucketUnits
}
