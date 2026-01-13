package datetime

import "time"

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.Local)
}
