package iis_api_entities

type Group struct {
	Id            int    `json:"id,omitempty" db:"id"`
	Name          string `json:"name,omitempty" db:"name"`
	FacultyId     int    `json:"facultyId,omitempty" db:"faculty_id"`
	SpreadsheetId string `json:"-" db:"spreadsheet_id"`
	AdminId       int    `json:"-" db:"admin_id"`
}
