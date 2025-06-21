package iis_api_entities

type (
	DayName        string
	LessonSchedule map[DayName][]Lesson
)

type ScheduleInfo struct {
	Schedule LessonSchedule `json:"schedules"`
}
