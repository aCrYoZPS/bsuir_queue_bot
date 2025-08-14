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
	DateTime       time.Time
}

func NewPersistedLesson(groupId int64, subgroupNumber entities.Subgroup, lessonType entities.LessonType, subject string, dateTime time.Time) *Lesson {
	return &Lesson{
		GroupId:        groupId,
		SubgroupNumber: subgroupNumber,
		LessonType:     lessonType,
		Subject:        subject,
		DateTime:       dateTime,
	}
}

func ToLessonEntity(lesson *Lesson) *entities.Lesson {
	return &entities.Lesson{
		GroupId:        lesson.GroupId,
		Subject:        lesson.Subject,
		SubgroupNumber: entities.Subgroup(lesson.SubgroupNumber),
		StartTime:      entities.TimeOnly(lesson.DateTime.AddDate(lesson.DateTime.Year(), int(lesson.DateTime.Month()), lesson.DateTime.Day())),
	}
}

func FromLessonEntity(lesson *entities.Lesson, date time.Time) *Lesson {
	return &Lesson{
		GroupId:        lesson.GroupId,
		LessonType:     lesson.LessonType,
		Subject:        lesson.Subject,
		SubgroupNumber: int8(lesson.SubgroupNumber),
		DateTime:       date.Add(time.Duration(time.Time(lesson.StartTime).UTC().Unix())),
	}
}
