package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
)

const (
	LESSONS_TABLE = "lessons"
	QUERY_TIMEOUT = 10 * time.Second
)

type LessonsRepository struct {
	db *sql.DB
}

func NewLessonsRepository(db *sql.DB) interfaces.LessonsRepository {
	repo := &LessonsRepository{
		db: db,
	}

	return repo
}

func (repo *LessonsRepository) AddRange(lessons []entities.Lesson) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	storedLessons := repo.getSortedLessons(lessons)
	for _, lesson := range storedLessons {
		query := fmt.Sprintf("INSERT INTO %s (group_id, subject, lesson_type, subgroup_number, date, time) values ($1,$2,$3,$4,$5,$6)", LESSONS_TABLE)
		tx.ExecContext(ctx, query, lesson.GroupId, lesson.Subject, lesson.LessonType, lesson.SubgroupNumber, lesson.Date, lesson.Time)
	}
	err = tx.Commit()
	return err
}

func (repo *LessonsRepository) GetAll(groupId int64) ([]persistance.Lesson, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	query := fmt.Sprintf("SELECT group_id, lesson_type, subgroup_number, date, time FROM %s WHERE group_id = %s LIMIT 4", LESSONS_TABLE, fmt.Sprint(groupId))
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 4)
	i := 0
	for rows.Next() {
		err = rows.Scan(&lessons[i].GroupId, &lessons[i].LessonType, &lessons[i].SubgroupNumber, &lessons[i].Date, &lessons[i].Time)
		if err != nil {
			return nil, err
		}
	}
	return lessons, nil
}

func (repo *LessonsRepository) GetNext(subject string, groupId int64) ([]persistance.Lesson, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	date := time.Now().UTC().Unix()
	query := fmt.Sprintf("SELECT group_id, lesson_type, subgroup_number, date, time FROM %s ORDER BY date-%[2]s WHERE date-%[2]s > 0 AND subject=%s AND group_id = %s LIMIT 4", LESSONS_TABLE, fmt.Sprint(date), subject, fmt.Sprint(groupId))
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 4)
	i := 0
	for rows.Next() {
		err = rows.Scan(&lessons[i].GroupId, &lessons[i].LessonType, &lessons[i].SubgroupNumber, &lessons[i].Date, &lessons[i].Time)
		if err != nil {
			return nil, err
		}
	}
	return lessons, nil
}

func (repo *LessonsRepository) getSortedLessons(labworks []entities.Lesson) []persistance.Lesson {
	if len(labworks) == 0 {
		return nil
	}
	lessons := make([]persistance.Lesson, 0, len(labworks)*10)
	lessons = append(lessons, *persistance.FromLessonEntity(&labworks[0], time.Time(labworks[0].StartDate)))
	for _, labwork := range labworks {
		startDate, endDate := time.Time(labwork.StartDate), time.Time(labwork.EndDate)
		currentDate := startDate
		for !currentDate.Equal(endDate) {
			currentDate = currentDate.Add(time.Hour * 24 * 7 * utils.CalculateWeeksDistance(labwork.WeekNumber, utils.CalculateWeek(startDate)))
			lessons = append(lessons, *persistance.FromLessonEntity(&labwork, currentDate))
		}
	}

	sort.Slice(lessons, func(i, j int) bool {
		if !lessons[i].Date.Equal(lessons[j].Date) {
			return lessons[i].Date.Before(lessons[j].Date)
		}
		return lessons[i].Time.Before(lessons[j].Time)
	})

	return lessons
}
