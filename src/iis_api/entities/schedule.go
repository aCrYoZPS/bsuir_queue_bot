package iis_api_entities

import "time"

type (
	DayName        string
	LessonSchedule map[DayName][]Lesson
)

var DayToName = map[time.Weekday]DayName{
	time.Monday:    "Понедельник",
	time.Tuesday:   "Вторник",
	time.Wednesday: "Среда",
	time.Thursday:  "Четверг",
	time.Friday:    "Пятница",
	time.Saturday:  "Суббота",
	time.Sunday:    "Воскресенье",
}

type ScheduleInfo struct {
	Schedule LessonSchedule `json:"schedules"`
}
