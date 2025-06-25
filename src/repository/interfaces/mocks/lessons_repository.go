package mocks

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type LessonsRepositoryMock struct {
	interfaces.LessonsRepository
}

func NewLessonsRepositoryMock() *LessonsRepositoryMock {
	return &LessonsRepositoryMock{}
}

//TODO. Too lasy to write for now. And also don't have a slightiest clue, if we even need them.