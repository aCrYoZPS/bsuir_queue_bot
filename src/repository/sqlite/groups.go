package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iisEntities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/mattn/go-sqlite3"
)

const GROUPS_TABLE = "groups"

var _ interfaces.GroupsRepository = (*GroupsRepository)(nil)

type GroupsRepository struct {
	db *sql.DB
}

func NewGroupsRepository(db *sql.DB) (interfaces.GroupsRepository, error) {
	repo := &GroupsRepository{
		db: db,
	}

	return repo, nil
}

func (repos *GroupsRepository) GetAll() ([]iisEntities.Group, error) {
	rows, err := repos.db.Query(fmt.Sprintf("SELECT * FROM %s", GROUPS_TABLE))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	groups := make([]iisEntities.Group, 0)
	for rows.Next() {
		g := iisEntities.Group{}
		err := rows.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId, &g.AdminId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (repos *GroupsRepository) Add(group *iisEntities.Group) error {
	_, err := repos.db.Exec(fmt.Sprintf("INSERT INTO %s (name, faculty_id, spreadsheet_id, admin_id) VALUES ($1, $2, $3, $4)", GROUPS_TABLE),
		group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId)
	if err != nil {
		return err
	}

	return nil
}

func (repos *GroupsRepository) AddRange(groups []iisEntities.Group) error {
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

func (repos *GroupsRepository) AddNonPresented(groups []iisEntities.Group) error {
	query := fmt.Sprintf("INSERT INTO %s (name, faculty_id, spreadsheet_id, admin_id) VALUES ($1, $2, $3, $4)", GROUPS_TABLE)
	for _, group := range groups {
		_, err := repos.db.Exec(query, group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId)
		if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode != sqlite3.ErrConstraintUnique {
			return err
		}
	}
	return nil
}

func (repos *GroupsRepository) GetById(id int) (*iisEntities.Group, error) {
	row := repos.db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE id=$1", GROUPS_TABLE), id)
	g := &iisEntities.Group{}

	err := row.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId, &g.AdminId)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (repos *GroupsRepository) GetByName(name string) (*iisEntities.Group, error) {
	row := repos.db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE name=$1", name), GROUPS_TABLE)
	group := &iisEntities.Group{}

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

func (repos *GroupsRepository) Update(group *iisEntities.Group) error {
	_, err := repos.db.Exec("UPDATE groups SET name=$1, faculty_id=$2, spreadsheet_id=$3, admin_id=$4 WHERE id=$5",
		group.Name, group.FacultyId, group.SpreadsheetId, group.AdminId, group.Id)
	return err
}

func (repos *GroupsRepository) DoesGroupExist(groupName string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE name=$1", GROUPS_TABLE)
	row := repos.db.QueryRow(query, groupName)
	if row.Err() != nil {
		if errors.Is(row.Err(), sql.ErrNoRows) {
			return false, nil
		}
		return false, row.Err()
	}
	return true, nil
}

func (repos *GroupsRepository) GetAdmins(groupName string) ([]entities.User, error) {
	query := fmt.Sprintf("SELECT us.id, us.tg_id, us.group_id, us.full_name FROM %s AS us INNER JOIN %s AS gr ON gr.id=us.group_id WHERE gr.name=$1", USERS_TABLE, GROUPS_TABLE)
	rows, err := repos.db.Query(query, groupName)
	if err != nil {
		return nil, err
	}
	users := make([]entities.User, 0, 4)
	for rows.Next() {
		user := &entities.User{}
		rows.Scan(&user.Id, &user.TgId, &user.GroupId, &user.FullName)
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		users = append(users, *user)
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users, nil
}
