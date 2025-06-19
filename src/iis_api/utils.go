package iis_api

import (
	"net/http"
)

func GetCurrentWeek() int {
	url := "https://iis.bsuir.by/api/v1/student-groups"

	resp, err := http.Get(url)
	if err != nil {
		return -1
	}

	defer resp.Body.Close()
	week_byte := make([]byte, 1)
	_, err = resp.Body.Read(week_byte)
	if err != nil {
		return -1
	}

	return int(week_byte[0] - '0')
}
