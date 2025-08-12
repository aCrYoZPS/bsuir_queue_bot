package sqlite

import (
	"context"
	"database/sql"
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

func NewGroupsRepository(db *sql.DB) (*GroupsRepository, error) {
	repo := &GroupsRepository{
		db: db,
	}

	return repo, nil
}

func (repos *GroupsRepository) GetAll(ctx context.Context) ([]iisEntities.Group, error) {
	rows, err := repos.db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", GROUPS_TABLE))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]iisEntities.Group, 0)
	for rows.Next() {
		g := iisEntities.Group{}
		err := rows.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (repos *GroupsRepository) Add(ctx context.Context, group *iisEntities.Group) error {
	_, err := repos.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, name, faculty_id, spreadsheet_id) VALUES ($1, $2, $3, $4)", GROUPS_TABLE),
		group.Id, group.Name, group.FacultyId, group.SpreadsheetId)
	if err != nil {
		return err
	}

	return nil
}

func (repos *GroupsRepository) AddRange(ctx context.Context, groups []iisEntities.Group) error {
	query := fmt.Sprintf("INSERT INTO %s (id, name, faculty_id, spreadsheet_id) VALUES ", GROUPS_TABLE)
	args := []any{}
	placeholders := []string{}

	for _, g := range groups {
		placeholders = append(placeholders, "(?, ?, ?, ?)")
		args = append(args, g.Id, g.Name, g.FacultyId, g.SpreadsheetId)
	}

	query += strings.Join(placeholders, ",")
	stmt, err := repos.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	return err
}

func (repos *GroupsRepository) AddNonPresented(ctx context.Context, groups []iisEntities.Group) error {
	query := fmt.Sprintf("INSERT INTO %s (id, name, faculty_id, spreadsheet_id) VALUES ($1, $2, $3, $4)", GROUPS_TABLE)
	for _, group := range groups {
		_, err := repos.db.ExecContext(ctx, query, group.Id, group.Name, group.FacultyId, group.SpreadsheetId)
		if err, ok := err.(sqlite3.Error); ok && err.ExtendedCode != sqlite3.ErrConstraintUnique && err.ExtendedCode != sqlite3.ErrConstraintPrimaryKey {
			return err
		}
	}
	return nil
}

func (repos *GroupsRepository) GetById(ctx context.Context, id int) (*iisEntities.Group, error) {
	row := repos.db.QueryRowContext(ctx, fmt.Sprintf("SELECT * FROM %s WHERE id=$1", GROUPS_TABLE), id)
	g := &iisEntities.Group{}

	err := row.Scan(&g.Id, &g.Name, &g.FacultyId, &g.SpreadsheetId)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (repos *GroupsRepository) GetByName(ctx context.Context, name string) (*iisEntities.Group, error) {
	row := repos.db.QueryRowContext(ctx, fmt.Sprintf("SELECT id, name, faculty_id, spreadsheet_id FROM %s WHERE name=$1", GROUPS_TABLE), name)
	group := &iisEntities.Group{}

	err := row.Scan(&group.Id, &group.Name, &group.FacultyId, &group.SpreadsheetId)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (repos *GroupsRepository) Delete(ctx context.Context, id int) error {
	_, err := repos.db.ExecContext(ctx, "DELETE FROM groups WHERE id=$1", id)
	return err
}

func (repos *GroupsRepository) Update(ctx context.Context, group *iisEntities.Group) error {
	_, err := repos.db.ExecContext(ctx, "UPDATE groups SET id=$1, name=$2, faculty_id=$3, spreadsheet_id=$4 WHERE id=$5",
		group.Id, group.Name, group.FacultyId, group.SpreadsheetId, group.Id)
	return err
}

func (repos *GroupsRepository) DoesGroupExist(ctx context.Context, groupName string) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE name=$1)", GROUPS_TABLE)
	row := repos.db.QueryRowContext(ctx, query, groupName)
	exists := false
	if row.Err() != nil {
		return false, row.Err()
	}
	err := row.Scan(&exists)
	return exists, err
}

func (repos *GroupsRepository) GetAdmins(ctx context.Context, groupName string) ([]entities.User, error) {
	query := fmt.Sprintf("SELECT us.id, us.tg_id, us.group_id, us.full_name FROM %s AS us INNER JOIN %s AS gr ON gr.id=us.group_id WHERE gr.name=$1", USERS_TABLE, GROUPS_TABLE)
	rows, err := repos.db.QueryContext(ctx, query, groupName)
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
