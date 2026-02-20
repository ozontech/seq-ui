package aggregation_ts

import (
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func GetIntervals(aggregations []*seqapi.AggregationQuery) ([]time.Duration, error) {
	aggIntervals := make([]time.Duration, 0, len(aggregations))
	for _, agg := range aggregations {
		if agg.Func != seqapi.AggFunc_AGG_FUNC_COUNT {
			aggIntervals = append(aggIntervals, 0)
			continue
		}

		interval, err := time.ParseDuration(*agg.Interval)
		if err != nil {
			return nil, err
		}

		aggIntervals = append(aggIntervals, interval)
	}

	return aggIntervals, nil
}
