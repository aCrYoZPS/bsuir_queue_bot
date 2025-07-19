package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

const GROUPS_TABLE = "groups"

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

func (repos *GroupsRepository) GetAll() ([]entities.Group, error) {
	rows, err := repos.db.Query(fmt.Sprintf("SELECT * FROM %s", GROUPS_TABLE))
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

func (repos *GroupsRepository) Add(group *entities.Group) error {
	_, err := repos.db.Exec(fmt.Sprintf("INSERT INTO %s (name, faculty_id, spreadsheet_id, admin_id) VALUES ($1, $2, $3, $4)", GROUPS_TABLE),
		group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId)
	if err != nil {
		return err
	}

	return nil
}

func (repos *GroupsRepository) AddRange(groups []entities.Group) error {
	query := fmt.Sprintf("INSERT INTO %s (name, faculty_id, spreadsheet_id, admin_id) VALUES ", GROUPS_TABLE)
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

func (repos *GroupsRepository) GetById(id int) (*entities.Group, error) {
	row := repos.db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE id=$1", GROUPS_TABLE), id)
	g := &entities.Group{}

	err := row.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId, &g.AdminId)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (repos *GroupsRepository) GetByName(name string) (*entities.Group, error) {
	row := repos.db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE name=$1", name), GROUPS_TABLE)
	group := &entities.Group{}

	err := row.Scan(&group.Id, &group.Name, &group.FacultyId, &group.SpreadsheetId, &group.AdminId)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (repos *GroupsRepository) Delete(id int) error {
	_, err := repos.db.Exec("DELETE FROM groups WHERE id=$1", id)
	return err
}

func (repos *GroupsRepository) Update(group *entities.Group) error {
	_, err := repos.db.Exec("UPDATE groups SET name=$1, faculty_id=$2, spreadsheet_id=$3, admin_id=$4 WHERE id=$5",
		group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId, group.Id)
	return err
}
