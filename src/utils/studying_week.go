package utils

import (
	"slices"
	"time"
)

var firstWeek = time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)

func CalculateWeek(date time.Time) int8 {
	duration := date.Sub(firstWeek)
	return int8(((duration/time.Hour)/24)%28)/7 + 1
}

type Week = int8

// Returns value from 0 to 3, to measure distance in weeks between labworks.
// Doesn't handle cases, where week is unpresent in slice of weeks
func CalculateWeeksDistance(weeks []Week, current Week) time.Duration {
	return time.Duration(weeks[(slices.Index(weeks, current)+1)%len(weeks)] - current)
}
