package mocks

import (
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
)

type LessonsRepositoryMock struct {
	interfaces.LessonsRepository
}

func NewLessonsRepositoryMock() *LessonsRepositoryMock {
	return &LessonsRepositoryMock{}
}

func (*LessonsRepositoryMock) GetAllLabworks(int64) ([]persistance.Lesson, error) {
	start, _ := time.Parse(time.DateOnly, "2025-02-15")
	end, _ := time.Parse(time.DateOnly, "2025-06-07")
	startTime, _ := time.Parse(time.TimeOnly, "09:00:00")
	secondStartTime, _ := time.Parse(time.TimeOnly, "10:35:00")
	return []persistance.Lesson{
		{
			Subject:        "ООП",
			LessonType:     "ЛР",
			SubgroupNumber: 0,
			Date:           start,
			Time:           startTime,
			GroupId:        0,
		},
		{
			Subject:        "AВС",
			LessonType:     "ЛР",
			SubgroupNumber: 0,
			Date:           end,
			Time:           secondStartTime,
			GroupId:        0,
		},
	}, nil
}

func (mock *LessonsRepositoryMock) GetNextLabworks(subject string, groupId int64) ([]persistance.Lesson, error) {
	return mock.GetAllLabworks(groupId)
}
