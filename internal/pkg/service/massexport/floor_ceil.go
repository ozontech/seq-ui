package massexport

import (
	"errors"
	"time"
)

var (
	errNotIntegerNumberOfMinutes = errors.New("part length is not an integer number of minutes")
	errNotDivisorOfTheDay        = errors.New("part length is not divisor of the day")
)

const day = 24 * time.Hour

func checkPartLength(length time.Duration) error {
	if length%time.Minute != 0 {
		return errNotIntegerNumberOfMinutes
	}

	if day%length != 0 {
		return errNotDivisorOfTheDay
	}

	return nil
}

func getMidnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func floor(t time.Time, partLength time.Duration) time.Time {
	base := getMidnight(t)
	a := t.Sub(base)

	return base.Add(a - a%partLength)
}

func ceil(t time.Time, partLength time.Duration) time.Time {
	base := getMidnight(t)
	a := t.Sub(base)

	if a%partLength == 0 {
		return base.Add(a)
	}

	return base.Add(a - a%partLength + partLength)
}
