package aggregationts

import "github.com/ozontech/seq-ui/pkg/seqapi/v1"

func GetIntervals(aggregations []*seqapi.AggregationQuery) []*string {
	aggIntervals := make([]*string, 0, len(aggregations))
	for _, agg := range aggregations {
		if agg.Func != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggIntervals = append(aggIntervals, nil)
			continue
		}
		aggIntervals = append(aggIntervals, agg.Interval)
	}

	return aggIntervals
}
