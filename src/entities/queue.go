package entities

import (
	"time"
)


type OrderField int8

const (
	BySubmission OrderField = iota + 1
	ByLabworkNumber
)

type OrderType struct {
	Ascending bool
	Value     OrderField
}

type Labwork struct {
	SubmitTime    time.Time
	User          User
	LessonId      int64
	LabworkNumber int8
}

type Queue []Labwork

func NewQueue(labworks []Labwork) *Queue {
	queue := Queue(labworks)
	return &queue
}
