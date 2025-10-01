package datetime

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DateOnly time.Time

type TimeOnly time.Time

type TimeWithSeconds time.Time

type DateTime time.Time

func (dt *DateOnly) UnmarshalJSON(json []byte) error {
	dateString := strings.Trim(string(json), `"`)
	if dateString == "null" {
		*dt = DateOnly{}
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
	*dt = (DateOnly)(dateVal)
	return nil
}

func (dt DateOnly) MarshalJSON() ([]byte, error) {
	a, err := json.Marshal(time.Time(dt).Format("02.01.2006"))
	return a, err
}

func (dt DateOnly) Format(s string) string {
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

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(dt).Format("02.01.2006 15:04:05"))
}

func (dt *DateTime) UnmarshalJSON(json []byte) error {
	dateString := strings.Trim(string(json), `"`)
	if dateString == "null" {
		*dt = DateTime{}
		return errors.New("null time field")
	}
	dateTime := strings.Split(dateString, " ")
	if len(dateTime) != 2 {
		return errors.New("date is not in format 12.02.2024 15:02:23")
	}

	date := strings.Split(dateTime[0], ".")
	if len(date) != 3 {
		return errors.New("date is not in format 12.02.2023 15:02:34")
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

	timeVal := strings.Split(dateTime[1], ":")
	if len(timeVal) != 3 {
		return errors.New("time is not in format 15:05:23")
	}
	hours, err := strconv.Atoi(timeVal[0])
	if err != nil {
		return errors.New("time is not in format 15:05:23")
	}
	minutes, err := strconv.Atoi(timeVal[1])
	if err != nil {
		return errors.New("time is not if format 14:05:43")
	}
	seconds, err := strconv.Atoi(timeVal[2])
	if err != nil {
		return fmt.Errorf("time is not in format 14:05:53")
	}
	dateVal := time.Date(year, time.Month(months), days, hours, minutes, seconds, 0, time.UTC)
	*dt = (DateTime)(dateVal)
	return nil
}
