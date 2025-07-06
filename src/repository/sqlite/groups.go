package sqlite

import (
	"database/sql"
	"strings"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type GroupsRepository struct {
	interfaces.GroupsRepository
	db *sql.DB
}

func NewGroupsRepository(db *sql.DB) (interfaces.GroupsRepository, error) {
	repo := &GroupsRepository{
		db: db,
	}

	return repo, nil
}

func (repos *GroupsRepository) GetAllGroups() ([]entities.Group, error) {
	rows, err := repos.db.Query("SELECT * FROM groups")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	groups := make([]entities.Group, 0)
	for rows.Next() {
		g := entities.Group{}
		err := rows.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId, &g.AdminId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (repos *GroupsRepository) AddGroup(group *entities.Group) error {
	_, err := repos.db.Exec("INSERT INTO groups (name, faculty_id, spreadsheet_id, admin_id) VALUES ($1, $2, $3, $4)",
		group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId)
	if err != nil {
		return err
	}

	return nil
}

func (repos *GroupsRepository) AddRange(groups []entities.Group) error {
	query := "INSERT INTO groups (name, faculty_id, spreadsheet_id, admin_id) VALUES "
	args := []any{}
	placeholders := []string{}

	for _, g := range groups {
		placeholders = append(placeholders, "(?, ?, ?, ?)")
		args = append(args, g.Name, g.FacultyId, g.SpreadsheetId, g.AdminId)
	}

	query += strings.Join(placeholders, ",")
	stmt, err := repos.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	return err
}

func (repos *GroupsRepository) GetGroup(id int) (*entities.Group, error) {
	row := repos.db.QueryRow("SELECT * FROM groups WHERE id=$1", id)
	g := &entities.Group{}
	err := row.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId, &g.AdminId)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (repos *GroupsRepository) DeleteGroup(id int) error {
	_, err := repos.db.Exec("DELETE FROM groups WHERE id=$1", id)
	if err != nil {
		return err
	}

	return nil
}
