package datetime

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

type DateTime time.Time

type TimeOnly time.Time

type TimeWithSeconds time.Time

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
	a, err :=  json.Marshal(time.Time(dt).Format("02.01.2006")) 
	return a, err
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
	return json.Marshal(time.Time(to).Format("15:04"))
}

func (to TimeOnly) Format(fmt string) string {
	return time.Time(to).Format(fmt)
}

func (to *TimeWithSeconds) UnmarshalJSON(json []byte) error {
	timeString := strings.Trim(string(json), `"`)
	if timeString == "null" {
		*to = TimeWithSeconds{}
		return errors.New("null time field")
	}
	layout := "15:04:05"
	timeVal, err := time.Parse(layout, timeString)
	*to = TimeWithSeconds(timeVal)
	return err
}

func (to TimeWithSeconds) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(to).Format("15:04:05"))
}

func (to TimeWithSeconds) Format(fmt string) string {
	return time.Time(to).Format(fmt)
}
