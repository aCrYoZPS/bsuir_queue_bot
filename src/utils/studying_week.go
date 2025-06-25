package utils

import "time"

var firstWeek = time.Date(2025, time.January, 13, 0, 0, 0, 0, time.UTC)

func CalculateWeek(date time.Time) int8 {
	duration := date.Sub(firstWeek)
	return int8(((duration / time.Hour)/24)%28)/7 + 1
}
