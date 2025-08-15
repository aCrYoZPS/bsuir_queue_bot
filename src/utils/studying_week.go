package utils

import (
	"slices"
	"time"
)

var firstWeek = time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)

func CalculateWeek(date time.Time) int8 {
	duration := date.Sub(firstWeek)
	if duration == 0 {
		duration = firstWeek.Sub(date)
	}
	week := int8(((duration/time.Hour)/24)%28)/7 + 1
	return week
}

type Week = int8

// Returns value from 1 to 4, to measure distance in weeks between labworks.
// Doesn't handle cases, where week is unpresent in slice of weeks
func CalculateWeeksDistance(weeks []Week, current Week) int8 {
	dist := (weeks[(slices.Index(weeks, current)+1)%len(weeks)] - current)
	if dist == 0 {
		return 4
	}
	return dist
}
