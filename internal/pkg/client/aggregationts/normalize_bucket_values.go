package aggregationts

import (
	"fmt"
	"time"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

func NormalizeBucketValues(aggregations []*seqapi.Aggregation, aggIntervals []*string) error {
	for i, agg := range aggregations {
		if agg == nil || agg.Buckets == nil || aggIntervals[i] == nil {
			continue
		}

		interval, err := time.ParseDuration(*aggIntervals[i])
		if err != nil {
			return fmt.Errorf("failed to parse aggregation interval: %w", err)
		}

		for _, bucket := range agg.Buckets {
			if bucket == nil || bucket.Value == nil {
				continue
			}

			*bucket.Value /= interval.Seconds()
		}
	}

	return nil
}
