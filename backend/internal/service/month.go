package service

import (
	"errors"
	"time"
	_ "time/tzdata"
)

var ErrInvalidTimezone = errors.New("invalid timezone")

func timezoneLocation(timezone string) (*time.Location, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, ErrInvalidTimezone
	}
	return location, nil
}

func monthRange(month, timezone string) (time.Time, time.Time, *time.Location, error) {
	location, err := timezoneLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, nil, err
	}
	if len(month) != len("2006-01") {
		return time.Time{}, time.Time{}, nil, ErrInvalidMonth
	}
	parsed, err := time.ParseInLocation("2006-01", month, location)
	if err != nil || parsed.Format("2006-01") != month {
		return time.Time{}, time.Time{}, nil, ErrInvalidMonth
	}
	start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, location)
	return start.UTC(), start.AddDate(0, 1, 0).UTC(), location, nil
}
