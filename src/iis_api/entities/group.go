package iis_api_entities

type Group struct {
	Id            int    `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	FacultyId     int    `json:"facultyId,omitempty"`
	SpreadsheetId string `json:"-" sql:"spreadsheet_id"`
}
