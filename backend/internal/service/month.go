package service

import (
	"time"
)

func monthRange(month string) (time.Time, time.Time, error) {
	if len(month) != len("2006-01") {
		return time.Time{}, time.Time{}, ErrInvalidMonth
	}
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidMonth
	}
	start = start.UTC()
	return start, start.AddDate(0, 1, 0), nil
}
