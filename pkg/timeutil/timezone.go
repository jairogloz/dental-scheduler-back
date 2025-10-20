package timeutil

import (
	"fmt"
	"time"
)

// ConvertTimeToTimezone converts a UTC time to the specified IANA timezone
// Returns an error if the timezone is invalid or empty
func ConvertTimeToTimezone(utcTime time.Time, timezone string) (time.Time, error) {
	if timezone == "" {
		return time.Time{}, fmt.Errorf("timezone cannot be empty")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}

	return utcTime.In(loc), nil
}

// ConvertTimesToTimezone converts start and end UTC times to the specified IANA timezone
// Returns an error if the timezone is invalid or empty
func ConvertTimesToTimezone(startUTC, endUTC time.Time, timezone string) (start, end time.Time, err error) {
	start, err = ConvertTimeToTimezone(startUTC, timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end, err = ConvertTimeToTimezone(endUTC, timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, end, nil
}
