package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type UsersRepository struct {
	interfaces.UsersRepository
	db *sql.DB
}

const (
	USERS_TABLE = "users"
	ROLES_TABLE = "users_roles"
)

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

func (repo *UsersRepository) GetById(id int64) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	query := fmt.Sprintf(`SELECT %[1]s.id, %[1]s.tg_id, %[1]s.group_id, %[1]s.full_name, %[3]s.name, %[2]s.role_name FROM %[1]s 
						INNER JOIN %[2]s ON %[1]s.id = %[2]s.user_id 
						INNER JOIN %[3]s ON %[1]s.group_id=%[3]s.id WHERE %[1]s.id = $1`, USERS_TABLE, ROLES_TABLE, GROUPS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	user := &entities.User{}
	for rows.Next() {
		var roleName string
		rows.Scan(&user.Id, &user.TgId, &user.GroupId, &user.FullName, &user.GroupName, &roleName)
		user.Roles = append(user.Roles, entities.RoleFromString(roleName))
	}
	if rows.Err() != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UsersRepository) GetByTgId(tgId int64) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	query := fmt.Sprintf("SELECT %[1]s.id, %[1]s.tg_id, %[1]s.group_id, %[1]s.full_name, %[2]s.role_name FROM %[1]s INNER JOIN %[2]s ON %[1]s.id = %[2]s.user_id WHERE %[1]s.tg_id = $1", USERS_TABLE, ROLES_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, tgId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user := &entities.User{}
	for rows.Next() {
		var roleName string
		err = rows.Scan(&user.Id, &user.TgId, &user.GroupId, &user.FullName, &roleName)
		if err != nil {
			return nil, err
		}
		user.Roles = append(user.Roles, entities.RoleFromString(roleName))
	}
	if rows.Err() != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UsersRepository) GetAll() ([]entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	query := fmt.Sprintf("SELECT %[1]s.id, %[1]s.tg_id, %[1]s.group_id, %[1]s.full_name, %[2]s.role_name FROM %[1]s INNER JOIN %[2]s ON %[1]s.id = %[2]s.user_id", USERS_TABLE, ROLES_TABLE)
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	users := make([]entities.User, 0)

	user := &entities.User{}
	var (
		previousId = int64(0)
		curId      = previousId
	)
	for rows.Next() {
		var roleName string
		rows.Scan(&user.Id, &user.TgId, &user.GroupId, &user.FullName, &roleName)
		user.Roles = append(user.Roles, entities.RoleFromString(roleName))
		if previousId != curId {
			users = append(users, *user)
		}
		previousId = curId
		curId = user.Id
	}
	if rows.Err() != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UsersRepository) Add(user *entities.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := fmt.Sprintf("SELECT id FROM %s WHERE name=$1", GROUPS_TABLE)
	row := tx.QueryRowContext(ctx, query, user.GroupName)
	if row.Err() != nil {
		return err
	}
	row.Scan(&user.GroupId)
	query = fmt.Sprintf("INSERT INTO %s (tg_id, group_id, full_name) values ($1, $2, $3) RETURNING id", USERS_TABLE)
	row = tx.QueryRowContext(ctx, query, user.TgId, user.GroupId, user.FullName)
	id := int64(0)
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	query = fmt.Sprintf("INSERT INTO %s (user_id, role_name) values ($1, $2)", ROLES_TABLE)
	tx.ExecContext(ctx, query, id, entities.Basic)
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (repo *UsersRepository) AddRange(users []entities.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	tx, err := repo.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}
	for _, user := range users {
		query := fmt.Sprintf("INSERT INTO %s (tg_id, group_id, full_name) values ($1, $2, $3) RETURNING id", USERS_TABLE)
		row := tx.QueryRowContext(ctx, query, user.TgId, user.GroupId, user.FullName)
		id := int64(0)
		err = row.Scan(&id)
		if err != nil {
			return err
		}
		query = fmt.Sprintf("INSERT INTO %s (user_id, role_name) values ($1, $2)", ROLES_TABLE)
		tx.ExecContext(ctx, query, id, entities.Basic)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (repo *UsersRepository) Update(user *entities.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf("UPDATE %s SET tg_id=$1, group_id=$2, full_name=$3 WHERE id=$4", USERS_TABLE)
	_, err = tx.ExecContext(ctx, query, user.TgId, user.GroupId, user.FullName, user.Id)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE id=$1", ROLES_TABLE)
	_, err = tx.ExecContext(ctx, query, user.Id)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("INSERT INTO %s (id,role_name) values ($1, $2)", ROLES_TABLE)
	for _, role := range user.Roles {
		_, err = tx.ExecContext(ctx, query, user.Id, role.ToString())
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (repo *UsersRepository) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1", USERS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, id)
	return err
}
