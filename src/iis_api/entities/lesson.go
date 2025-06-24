package iis_api_entities

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// I mean, it's basically a enum.
type Subgroup int8

const (
	AllSubgroups   = Subgroup(0)
	FirstSubgroup  = Subgroup(1)
	SecondSubgroup = Subgroup(2)
)

type DateTime time.Time

type TimeOnly time.Time

// I'm not dumb, it uses the freest time
var timeFormat = "02.01.2006 +0300"

type Lesson struct {
	Subject        string   `json:"subject,omitempty"`
	LessonType     string   `json:"lessonTypeAbbrev,omitempty"`
	SubgroupNumber Subgroup `json:"numSubgroup,omitempty"`
	WeekNumber     []int8   `json:"weekNumber,omitempty"`
	StartDate      DateTime `json:"startLessonDate"`
	StartTime      TimeOnly `json:"startLessonTime"`
	EndDate        DateTime `json:"endLessonDate"`
}

func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		t, _ := time.Parse(timeFormat, "01.01.1970")
		*dt = DateTime(t)
		return nil
	}
	t, err := time.Parse(timeFormat, s)
	if err != nil {
		return err
	}

	*dt = DateTime(t)

	return nil
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(dt))
}

func (dt DateTime) Format(s string) string {
	return time.Time(dt).Format(s)
}

func (to TimeOnly) UnmarshalJSON(bytes []byte) error {
	timeString := strings.Trim(string(bytes), `"`)
	if timeString == "null" {
		to = TimeOnly{}
		return errors.New("null time field")
	}
	timeVal, err := time.Parse(time.TimeOnly, timeString)
	if err != nil {
		return err
	}
	to = TimeOnly(timeVal)
	return nil
}

func (to TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(to))
}

func (to TimeOnly) Format(fmt string) string {
	return time.Time(to).Format(fmt)
}
