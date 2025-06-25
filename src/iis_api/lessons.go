package iis_api

import (
	"encoding/json"
	"os"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

func GetNextLessons(subject string, group_id string, subgroup entities.Subgroup, date time.Time) ([]entities.Lesson, error) {
	schedule_content, err := os.ReadFile("353502.json")
	if err != nil {
		return nil, err
	}

	var schedule_info entities.ScheduleInfo
	err = json.Unmarshal(schedule_content, &schedule_info)
	if err != nil {
		return nil, err
	}

	week := GetWeekForDate(date)

	return make([]entities.Lesson, 3), nil
}
