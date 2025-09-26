package entities

import "time"

type LessonRequest struct {
	SubmitTime    time.Time
	Id            int64
	LessonId      int64
	UserId        int64
	MsgId         int64
	ChatId        int64
	LabworkNumber int8
}

func NewLessonRequest(LessonId, UserId, MsgId, ChatId int64, LabworkNumber int8) *LessonRequest {
	return &LessonRequest{LessonId: LessonId, UserId: UserId, MsgId: MsgId, ChatId: ChatId, LabworkNumber: LabworkNumber, SubmitTime: time.Now()}
}
