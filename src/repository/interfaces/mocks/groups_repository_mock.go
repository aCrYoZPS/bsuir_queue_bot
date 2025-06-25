package mocks

import (
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

var MockGroups = []iis_api_entities.Group{{Id: 0, Name: "353502", FacultyId: 5, SpreadsheetId: ""},
	{Id: 1, Name: "353503", FacultyId: 5, SpreadsheetId: ""}}

type GroupsRepositoryMock struct {
	interfaces.GroupsRepository
}

func NewGroupsRepositoryMock() *GroupsRepositoryMock {
	return &GroupsRepositoryMock{}
}

func (*GroupsRepositoryMock) GetAllGroups() ([]iis_api_entities.Group, error) {
	return MockGroups, nil
}

func (*GroupsRepositoryMock) SaveAllGroups([]iis_api_entities.Group) error {
	return nil
}
