package sqlite

import (
	"database/sql"
	"strings"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
)

type LessonsRepository struct {
	db *sql.DB
}

func NewLessonsRepository(db *sql.DB) (interfaces.LessonsRepository, error) {
	repo := &LessonsRepository{
		db: db,
	}

	_, err := repo.db.Exec(`CREATE TABLE IF NOT EXISTS lessons
							(
								id INTEGER PRIMARY KEY,
								subject TEXT UNIQUE,
								lesson_type TEXT,
								subgroup_number INTEGER,
								week_numbers TEXT,
								start_date INTEGER,
								start_time INTEGER,
								end_date INTEGER,
								group_id INTEGER
							)`,
	)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (repos *LessonsRepository) GetAll() ([]entities.Lesson, error) {
	rows, err := repos.db.Query("SELECT * FROM lessons")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var weekNumbers string
	var startDate int64
	var startTime int64
	var endDate int64

	lessons := make([]entities.Lesson, 0)
	for rows.Next() {
		l := entities.Lesson{}
		err := rows.Scan(&l.Id, &l.Subject, &l.LessonType, &l.SubgroupNumber,
			&weekNumbers, &startDate, &startTime, &endDate, &l.GroupId)
		if err != nil {
			return nil, err
		}

		weekNumbersArray, err := utils.ParseArray(weekNumbers)
		if err != nil {
			return nil, err
		}

		l.WeekNumber = weekNumbersArray
		l.StartDate = entities.DateTime(time.Unix(startDate, 0))
		l.StartTime = entities.TimeOnly(time.Unix(startTime, 0))
		l.EndDate = entities.DateTime(time.Unix(endDate, 0))

		lessons = append(lessons, l)
	}

	return lessons, nil
}

func (repos *LessonsRepository) GetAllByGroupId(groupId int) ([]entities.Lesson, error) {
	rows, err := repos.db.Query("SELECT * FROM lessons where groupId=$1", groupId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var weekNumbers string
	var startDate int64
	var startTime int64
	var endDate int64

	lessons := make([]entities.Lesson, 0)
	for rows.Next() {
		l := entities.Lesson{}
		err := rows.Scan(&l.Id, &l.Subject, &l.LessonType, &l.SubgroupNumber,
			&weekNumbers, &startDate, &startTime, &endDate, &l.GroupId)
		if err != nil {
			return nil, err
		}

		weekNumbersArray, err := utils.ParseArray(weekNumbers)
		if err != nil {
			return nil, err
		}

		l.WeekNumber = weekNumbersArray
		l.StartDate = entities.DateTime(time.Unix(startDate, 0))
		l.StartTime = entities.TimeOnly(time.Unix(startTime, 0))
		l.EndDate = entities.DateTime(time.Unix(endDate, 0))

		lessons = append(lessons, l)
	}

	return lessons, nil
}

func (repos *LessonsRepository) Add(lesson *entities.Lesson) error {
	_, err := repos.db.Exec(`INSERT INTO lessons (subject,lesson_type,subgroup_number,week_numbers,
												  start_date,start_time,end_date,group_id)
										  VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		lesson.Subject, lesson.LessonType, lesson.SubgroupNumber,
		utils.ArrayToString(lesson.WeekNumber), time.Time(lesson.StartDate).Unix(),
		time.Time(lesson.StartTime).Unix(), time.Time(lesson.EndDate).Unix(), lesson.GroupId)
	if err != nil {
		return err
	}

	return nil
}

func (repos *LessonsRepository) AddRange(lessons []entities.Lesson) error {
	query := `INSERT INTO lessons (subject,lesson_type,subgroup_number,week_numbers,
								   start_date,start_time,end_date,group_id) VALUES `
	args := []any{}
	placeholders := []string{}

	for _, l := range lessons {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, l.Subject, l.LessonType, l.SubgroupNumber, utils.ArrayToString(l.WeekNumber),
			time.Time(l.StartDate).Unix(), time.Time(l.StartTime).Unix(), time.Time(l.EndDate).Unix(), l.GroupId)
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

func (repos *LessonsRepository) GetById(id int) (*entities.Lesson, error) {
	var weekNumbers string
	var startDate int64
	var startTime int64
	var endDate int64
	l := &entities.Lesson{}

	row := repos.db.QueryRow("SELECT * FROM lessons WHERE id=$1", id)
	err := row.Scan(&l.Id, &l.Subject, &l.LessonType, &l.SubgroupNumber,
		&weekNumbers, &startDate, &startTime, &endDate, &l.GroupId)
	if err != nil {
		return nil, err
	}

	weekNumbersArray, err := utils.ParseArray(weekNumbers)
	if err != nil {
		return nil, err
	}

	l.WeekNumber = weekNumbersArray
	l.StartDate = entities.DateTime(time.Unix(startDate, 0))
	l.StartTime = entities.TimeOnly(time.Unix(startTime, 0))
	l.EndDate = entities.DateTime(time.Unix(endDate, 0))

	return l, nil
}

func (repos *LessonsRepository) Delete(id int) error {
	_, err := repos.db.Exec("DELETE FROM lessons WHERE id=$1", id)
	return err
}

func (repos *LessonsRepository) Update(lesson *entities.Lesson) error {
	_, err := repos.db.Exec(`UPDATE lessons SET subject=$1, lesson_type=$2, subgroup_number=$3, week_numbers=$4,
												start_date=$5, start_time=$6, end_date=$7, group_id=$8 WHERE id=$9`,
		lesson.Subject, lesson.LessonType, lesson.SubgroupNumber, utils.ArrayToString(lesson.WeekNumber),
		time.Time(lesson.StartDate).Unix(), time.Time(lesson.StartTime).Unix(), time.Time(lesson.EndDate),
		lesson.GroupId, lesson.Id)

	return err
}
