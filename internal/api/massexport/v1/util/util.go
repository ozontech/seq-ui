package util

import (
	"fmt"
	"time"
)

const defaultWindow = 10 * time.Second

func ParseWindow(s string) (time.Duration, error) {
	if s == "" {
		return defaultWindow, nil
	}

	window, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}

	switch {
	case window == 0:
		return 0, fmt.Errorf("zero window: %s", s)
	case window < 0:
		return 0, fmt.Errorf("negative window: %s", s)
	default:
		return window, nil
	}
}
