package iis_api_entities

type Lesson struct {
	Subject        string
	LessonType     string
	SubgroupNumber int8
	WeekNumber     []int8
}
