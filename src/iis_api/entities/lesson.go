package iis_api_entities

import (
	datetime "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/date_time"
)

type Subgroup = int8

const (
	AllSubgroups   = Subgroup(0)
	FirstSubgroup  = Subgroup(1)
	SecondSubgroup = Subgroup(2)
)

type LessonType = string

const (
	Lecture LessonType = "ЛК"
	Seminar LessonType = "ПЗ"
	Labwork LessonType = "ЛР"
)

type Lesson struct {
	Id             int               `json:"-" db:"id"`
	Subject        string            `json:"subject,omitempty" db:"subject"`
	LessonType     LessonType        `json:"lessonTypeAbbrev,omitempty" db:"lesson_type"`
	SubgroupNumber Subgroup          `json:"numSubgroup,omitempty" db:"subgroup_number"`
	WeekNumber     []int8            `json:"weekNumber,omitempty" db:"week_number"`
	StartDate      datetime.DateOnly `json:"startLessonDate" db:"start_date"`
	StartTime      datetime.TimeOnly `json:"startLessonTime" db:"start_time"`
	EndDate        datetime.DateOnly `json:"endLessonDate" db:"end_date"`
	GroupId        int64             `json:"-" db:"group_id"`
}

const AVAILABLE_LESSONS_COUNT = 4
