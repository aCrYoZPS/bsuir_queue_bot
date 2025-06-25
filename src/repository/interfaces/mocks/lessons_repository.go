package mocks

import (
	"time"

	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type LessonsRepositoryMock struct {
	interfaces.LessonsRepository
}

func NewLessonsRepositoryMock() *LessonsRepositoryMock {
	return &LessonsRepositoryMock{}
}

func (*LessonsRepositoryMock) GetAllLabworks(*iis_api_entities.Group) ([]iis_api_entities.Lesson, error) {
	start, _ := time.Parse(time.DateOnly, "2025-02-15")
	end, _ := time.Parse(time.DateOnly, "2025-06-07")
	startTime, _ := time.Parse(time.TimeOnly, "09:00:00")
	secondStartTime, _ := time.Parse(time.TimeOnly, "10:35:00")
	return []iis_api_entities.Lesson{
		{
			Subject:        "ООП",
			LessonType:     "ЛР",
			SubgroupNumber: 0,
			WeekNumber:     []int8{1, 3},
			StartDate:      iis_api_entities.DateTime(start),
			StartTime:      iis_api_entities.TimeOnly(secondStartTime),
			EndDate:        iis_api_entities.DateTime(end),
		},
		{
			Subject:        "AВС",
			LessonType:     "ЛР",
			SubgroupNumber: 0,
			WeekNumber:     []int8{1, 3},
			StartDate:      iis_api_entities.DateTime(start),
			StartTime:      iis_api_entities.TimeOnly(startTime),
			EndDate:        iis_api_entities.DateTime(end),
		},
	}, nil
}
