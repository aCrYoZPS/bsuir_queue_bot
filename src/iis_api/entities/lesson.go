package iis_api_entities

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Subgroup = int8

const (
	AllSubgroups   = Subgroup(0)
	FirstSubgroup  = Subgroup(1)
	SecondSubgroup = Subgroup(2)
)

type DateTime time.Time

type TimeOnly time.Time

type LessonType = string

const (
	Lecture LessonType = "ЛК"
	Seminar LessonType = "ПЗ"
	Labwork LessonType = "ЛР"
)

type Lesson struct {
	Id             int        `json:"-" db:"id"`
	Subject        string     `json:"subject,omitempty" db:"subject"`
	LessonType     LessonType `json:"lessonTypeAbbrev,omitempty" db:"lesson_type"`
	SubgroupNumber Subgroup   `json:"numSubgroup,omitempty" db:"subgroup_number"`
	WeekNumber     []int8     `json:"weekNumber,omitempty" db:"week_number"`
	StartDate      DateTime   `json:"startLessonDate" db:"start_date"`
	StartTime      TimeOnly   `json:"startLessonTime" db:"start_time"`
	EndDate        DateTime   `json:"endLessonDate" db:"end_date"`
	GroupId        int64      `json:"-" db:"group_id"`
}

const AVAILABLE_LESSONS_COUNT = 4

func (dt *DateTime) UnmarshalJSON(json []byte) error {
	dateString := strings.Trim(string(json), `"`)
	if dateString == "null" {
		*dt = DateTime{}
		return errors.New("null time field")
	}
	date := strings.Split(dateString, ".")
	if len(date) != 3 {
		return errors.New("date is not in format 12.02.2023")
	}
	days, err := strconv.Atoi(date[0])
	if err != nil {
		return err
	}
	months, err := strconv.Atoi(date[1])
	if err != nil {
		return err
	}
	year, err := strconv.Atoi(date[2])
	if err != nil {
		return err
	}
	dateVal := time.Date(year, time.Month(months), days, 0, 0, 0, 0, time.UTC)
	*dt = (DateTime)(dateVal)
	return nil
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(dt))
}

func (dt DateTime) Format(s string) string {
	return time.Time(dt).Format(s)
}

func (to *TimeOnly) UnmarshalJSON(json []byte) error {
	timeString := strings.Trim(string(json), `"`)
	if timeString == "null" {
		*to = TimeOnly{}
		return errors.New("null time field")
	}
	layout := "15:04"
	timeVal, err := time.Parse(layout, timeString)
	*to = TimeOnly(timeVal)
	return err
}

func (to TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(to))
}

func (to TimeOnly) Format(fmt string) string {
	return time.Time(to).Format(fmt)
}
