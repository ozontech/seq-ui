package massexport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckPartLength(t *testing.T) {
	type TestCase struct {
		length time.Duration
		err    error
	}

	tests := []TestCase{
		{60 * time.Minute, nil},
		{40 * time.Minute, nil},
		{30 * time.Minute, nil},
		{20 * time.Minute, nil},
		{15 * time.Minute, nil},

		{1*time.Hour + 30*time.Second, errNotIntegerNumberOfMinutes},
		{30 * time.Second, errNotIntegerNumberOfMinutes},
		{61 * time.Second, errNotIntegerNumberOfMinutes},

		{50 * time.Minute, errNotDivisorOfTheDay},
		{65 * time.Minute, errNotDivisorOfTheDay},
		{70 * time.Minute, errNotDivisorOfTheDay},
		{110 * time.Minute, errNotDivisorOfTheDay},
	}

	for _, test := range tests {
		err := checkPartLength(test.length)
		if test.err == nil {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, test.err)
		}
	}
}

func TestFloor(t *testing.T) {
	type TestCase struct {
		valueStr, expectedStr string
	}

	tests := []TestCase{
		{valueStr: "2024-10-10T10:00:00+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T10:00:00.001+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T10:00:01+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T10:20:00+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T10:40:00+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T11:00:00+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
	}

	for _, test := range tests {
		value, err := time.Parse(time.RFC3339, test.valueStr)
		require.NoError(t, err)
		expected, err := time.Parse(time.RFC3339, test.expectedStr)
		require.NoError(t, err)

		require.Equal(t, expected, floor(value, defaultPartLength))
	}
}

func TestCeil(t *testing.T) {
	type TestCase struct {
		valueStr, expectedStr string
	}

	tests := []TestCase{
		{valueStr: "2024-10-10T10:00:00+03:00", expectedStr: "2024-10-10T10:00:00+03:00"},
		{valueStr: "2024-10-10T10:20:00+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
		{valueStr: "2024-10-10T10:40:00+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
		{valueStr: "2024-10-10T10:59:59+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
		{valueStr: "2024-10-10T10:59:59.999+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
		{valueStr: "2024-10-10T11:00:00+03:00", expectedStr: "2024-10-10T11:00:00+03:00"},
	}

	for _, test := range tests {
		value, err := time.Parse(time.RFC3339, test.valueStr)
		require.NoError(t, err)
		expected, err := time.Parse(time.RFC3339, test.expectedStr)
		require.NoError(t, err)

		require.Equal(t, expected, ceil(value, defaultPartLength))
	}
}
