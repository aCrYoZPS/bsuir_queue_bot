package utils

import (
	"slices"
	"time"
)

const (
	DaysInWeekCycle = 28
	LastWeek        = 4
)

var firstWeek = time.Date(2026, time.August, 3, 0, 0, 0, 0, time.Local)

func CalculateWeek(date time.Time) int8 {
	duration := date.Sub(firstWeek)
	if duration < 0 {
		duration = -duration
	}
	week := int8(((duration/time.Hour)/24)%DaysInWeekCycle)/7 + 1
	return week
}

type Week = int8

// Returns value from 1 to 4, to measure distance in weeks between labworks.
// Doesn't handle cases, where week is unpresent in slice of weeks
func CalculateWeeksDistance(weeks []Week, current Week) int8 {
	dist := (weeks[(slices.Index(weeks, current)+1)%len(weeks)] - current)
	if dist == 0 {
		return LastWeek
	}
	if dist < 0 {
		return -dist
	}
	return dist
}
