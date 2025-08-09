package persistance

import (
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
)

type Lesson struct {
	Id             int64
	GroupId        int64
	LessonType     entities.LessonType
	Subject        string
	SubgroupNumber entities.Subgroup
	Date           time.Time
	Time           time.Time
}

func NewPersistedLesson(groupId int64, subgroupNumber entities.Subgroup, lessonType entities.LessonType, subject string, date time.Time, time time.Time) *Lesson {
	return &Lesson{
		GroupId:        groupId,
		SubgroupNumber: subgroupNumber,
		LessonType:     lessonType,
		Subject:        subject,
		Date:           date,
		Time:           time,
	}
}

func ToLessonEntity(lesson *Lesson) *entities.Lesson {
	return &entities.Lesson{
		GroupId:        lesson.GroupId,
		Subject:        lesson.Subject,
		SubgroupNumber: entities.Subgroup(lesson.SubgroupNumber),
		StartTime:      entities.TimeOnly(lesson.Time),
	}
}

func FromLessonEntity(lesson *entities.Lesson, date time.Time) *Lesson {
	return &Lesson{
		GroupId:        lesson.GroupId,
		LessonType:     lesson.LessonType,
		Subject:        lesson.Subject,
		SubgroupNumber: int8(lesson.SubgroupNumber),
		Time:           time.Time(lesson.StartTime),
		Date:           date,
	}
}
