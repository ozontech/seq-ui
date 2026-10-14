package api_error

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrQueryTooHeavy  = errors.New("query too heavy, try decreasing date range or make search query more precise")
	ErrTooManyBuckets = errors.New("aggregation has too many buckets, try decreasing interval")
	ErrOffsetConflict = errors.New(`only one of "offset" and "offset_id" must be provided`)
)

func CheckSearchLimit(limit, maximum int32) error {
	if limit == 0 || limit > maximum {
		return fmt.Errorf("'limit' must be greater than 0 and less than %v", maximum)
	}
	return nil
}

func CheckSearchOffsetLimit(offset, maximum int32) error {
	if offset > maximum {
		return ErrQueryTooHeavy
	}
	return nil
}

func CheckForOffsetConflict(offset int32, offsetID string) error {
	if offset != 0 && offsetID != "" {
		return ErrOffsetConflict
	}
	return nil
}

func CheckAggregationsCount(count, maximum int) error {
	if count > maximum {
		return fmt.Errorf("too many aggregations requested, limit is %v aggregations per request", maximum)
	}
	return nil
}

func CheckAggregationTsInterval(interval string, from, to time.Time, maximum int) error {
	i, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("invalid aggregation interval: %w", err)
	}
	count := int(to.Sub(from) / i)
	if count > maximum {
		return ErrTooManyBuckets
	}
	return nil
}
