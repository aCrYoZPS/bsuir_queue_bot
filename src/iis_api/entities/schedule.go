package iis_api_entities

type (
	DayName  string
	Schedule map[DayName][]Lesson
)
