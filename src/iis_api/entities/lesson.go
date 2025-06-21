package iis_api_entities

import "time"

// I mean, it's basically a enum.
type Subgroup int8

const (
	AllSubgroups   = Subgroup(0)
	FirstSubgroup  = Subgroup(1)
	SecondSubgroup = Subgroup(2)
)

type Lesson struct {
	Subject        string    `json:"subject,omitempty"`
	LessonType     string    `json:"lessonTypeAbbrev,omitempty"`
	SubgroupNumber Subgroup  `json:"numSubgroup,omitempty"`
	WeekNumber     []int8    `json:"weekNumber,omitempty"`
	StartDate      time.Time `json:"startLessonDate"`
	EndDate        time.Time `json:"endLessonDate"`
}
