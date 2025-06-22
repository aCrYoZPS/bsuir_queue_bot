package utils

import "time"

var firstWeek = time.Date(2025, time.January, 12, 0, 0, 0, 0, time.UTC)

func CalculateWeek(date time.Time) int8 {
	duration := date.Sub(firstWeek)
	return int8((int(duration.Hours()) / 24) % 28)
}
