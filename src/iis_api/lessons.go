package iis_api

import (
	_ "encoding/json"
	_ "os"
	_ "time"

	_ "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	_ "github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
)


//That'll probably be turned into service class, which contains injected repository. Commented it out to compile project, for now

// func GetNextLessons(subject string, group_id string, subgroup entities.Subgroup, date time.Time) ([]entities.Lesson, error) {
// 	schedule_content, err := os.ReadFile("353502.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var schedule_info entities.ScheduleInfo
// 	err = json.Unmarshal(schedule_content, &schedule_info)
// 	if err != nil {
// 		return nil, err
// 	}

// 	week := utils.CalculateWeek(date)


// 	return make([]entities.Lesson, 3), nil
// }
