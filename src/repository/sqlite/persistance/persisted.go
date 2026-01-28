package persistance

import (
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	_ "github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iisEntities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	datetime "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/date_time"
)

type Lesson struct {
	Id             int64
	GroupId        int64
	LessonType     iisEntities.LessonType
	Subject        string
	SubgroupNumber iisEntities.Subgroup
	DateTime       time.Time
}

func NewPersistedLesson(groupId int64, subgroupNumber iisEntities.Subgroup, lessonType iisEntities.LessonType, subject string, dateTime time.Time) *Lesson {
	return &Lesson{
		GroupId:        groupId,
		SubgroupNumber: subgroupNumber,
		LessonType:     lessonType,
		Subject:        subject,
		DateTime:       dateTime,
	}
}

func ToLessonEntity(lesson *Lesson) *iisEntities.Lesson {
	return &iisEntities.Lesson{
		GroupId:        lesson.GroupId,
		Subject:        lesson.Subject,
		SubgroupNumber: iisEntities.Subgroup(lesson.SubgroupNumber),
		StartTime:      datetime.TimeOnly(lesson.DateTime.AddDate(lesson.DateTime.Year(), int(lesson.DateTime.Month()), lesson.DateTime.Day())),
	}
}

func FromLessonEntity(lesson *iisEntities.Lesson, date time.Time) *Lesson {
	return &Lesson{
		GroupId:        lesson.GroupId,
		LessonType:     lesson.LessonType,
		Subject:        lesson.Subject,
		SubgroupNumber: int8(lesson.SubgroupNumber),
		DateTime:       date.Add(time.Time(lesson.StartTime).Sub(time.Time{})).AddDate(1, 0, 1),
	}
}

type OrderField = int8

const (
	BySubmission OrderField = iota
	ByLabworkNumber
)

type OrderType struct {
	Ascending bool
	Value     OrderField
}

type PersistedQueue struct {
	LessonId int64
	//Array, which specifies ordering by multiple fields. Order of application is the same as elements in array
	OrderedBy []OrderType
}

func FromOrderTypes(lessonId int64, orderTypes []entities.OrderType) *PersistedQueue {
	queue := &PersistedQueue{LessonId: lessonId}
	persistedTypes := make([]OrderType, 0, len(orderTypes))
	for _, orderType := range orderTypes {
		persistedTypes = append(persistedTypes, ToPersistentType(orderType))
	}
	queue.OrderedBy = persistedTypes
	return queue
}

func ToPersistentType(entity entities.OrderType) OrderType {
	return OrderType{Value: OrderField(entity.Value), Ascending: entity.Ascending}
}

func NewPersistedQueue(opts ...func(*PersistedQueue)) *PersistedQueue {
	queue := &PersistedQueue{}
	for _, opt := range opts {
		opt(queue)
	}
	if len(queue.OrderedBy) == 0 {
		queue.OrderedBy = []OrderType{{
			Ascending: true,
			Value:     BySubmission,
		}}
	}
	return queue
}

func WithSubmissionOrder(ascending bool) func(*PersistedQueue) {
	return func(pq *PersistedQueue) {
		pq.OrderedBy = append(pq.OrderedBy, OrderType{Value: BySubmission, Ascending: ascending})
	}
}

func WithLabworkOrder(ascending bool) func(*PersistedQueue) {
	return func(pq *PersistedQueue) {
		pq.OrderedBy = append(pq.OrderedBy, OrderType{Value: ByLabworkNumber, Ascending: ascending})
	}
}
