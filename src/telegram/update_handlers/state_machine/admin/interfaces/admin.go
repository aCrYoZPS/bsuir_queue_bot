package interfaces

type LessonsService interface {
	AddGroupLessons(groupName string) (url string, err error)
}