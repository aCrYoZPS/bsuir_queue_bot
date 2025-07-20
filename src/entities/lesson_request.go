package entities

type LessonRequest struct {
	Id       int64
	LessonId int64
	UserId   int64
}

func NewLessonRequest(LessonId, UserId int64) *LessonRequest {
	return &LessonRequest{LessonId: LessonId, UserId: UserId}
}
