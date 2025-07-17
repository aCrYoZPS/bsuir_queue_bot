package iis_api

import (
	"io"
	"net/http"
	"time"
)

var (
	reference_date      = time.Date(2025, time.June, 23, 0, 0, 0, 0, time.UTC)
	reference_date_week = 4
)

var weekday_map = map[time.Weekday]int{
	time.Monday:    0 * 24,
	time.Tuesday:   1 * 24,
	time.Wednesday: 2 * 24,
	time.Thursday:  3 * 24,
	time.Friday:    4 * 24,
	time.Saturday:  5 * 24,
	time.Sunday:    6 * 24,
}

func GetCurrentWeek() (int, error) {
	url := "https://iis.bsuir.by/api/v1/schedule/current-week"

	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()
	week_byte, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	return int(week_byte[0] - '0'), nil
}

func GetWeekForDate(date time.Time) int {
	date = date.Add(time.Duration(-weekday_map[date.Weekday()]) * time.Hour)
	weeks_delta := int(reference_date.Sub(date).Abs().Hours() / (24 * 7))
	week_number := 1
	if reference_date.Before(date) {
		week_number = (reference_date_week + weeks_delta%4) % 4
	} else {
		week_number = (reference_date_week - weeks_delta%4) % 4
	}
	if week_number == 0 {
		week_number = 4
	}
	return week_number
}
