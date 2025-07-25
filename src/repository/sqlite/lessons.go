package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"time"

	entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/utils"
	datastructures "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/data_structures"
)

const (
	LESSONS_TABLE = "lessons"
	QUERY_TIMEOUT = 10 * time.Second
)

var _ interfaces.LessonsRepository = (*LessonsRepository)(nil)

type LessonsRepository struct {
	interfaces.LessonsRepository
	db *sql.DB
}

func NewLessonsRepository(db *sql.DB) interfaces.LessonsRepository {
	repo := &LessonsRepository{
		db: db,
	}
	return repo
}

const savedFormat = time.RFC3339

func (repo *LessonsRepository) AddRange(lessons []*entities.Lesson) error {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	tx, err := repo.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}
	storedLessons := repo.getSortedLessons(lessons)
	for _, lesson := range storedLessons {
		storedDate := lesson.Date.Format(savedFormat)
		storedTime := lesson.Time.Format(savedFormat)
		query := fmt.Sprintf("INSERT INTO %s (group_id, subject, lesson_type, subgroup_number, date, time) values ($1,$2,$3,$4,$5,$6)", LESSONS_TABLE)
		tx.ExecContext(ctx, query, lesson.GroupId, lesson.Subject, lesson.LessonType, lesson.SubgroupNumber, storedDate, storedTime)
	}
	err = tx.Commit()
	return err
}

func (repo *LessonsRepository) GetAll(groupName string) ([]persistance.Lesson, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	query := fmt.Sprintf("SELECT group_id, lesson_type, subject, subgroup_number, date, time FROM %s WHERE $1 in (SELECT name from %s)", LESSONS_TABLE, GROUPS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 0, 100)
	i := 0
	var storedTime, storedDate = "", ""
	for rows.Next() {
		lesson := &persistance.Lesson{}
		err = rows.Scan(&lesson.GroupId, &lesson.LessonType, &lesson.Subject, &lesson.SubgroupNumber, &storedDate, &storedTime)
		if err != nil {
			return nil, err
		}
		lesson.Date, err = time.Parse(savedFormat, storedDate)
		if err != nil {
			return nil, err
		}
		lesson.Time, err = time.Parse(savedFormat, storedTime)
		if err != nil {
			return nil, err
		}
		i++
		lessons = append(lessons, *lesson)
	}
	if i == 0 {
		return nil, nil
	}
	return lessons, nil
}

func (repo *LessonsRepository) GetNext(subject string, groupId int64) ([]persistance.Lesson, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()

	date := time.Now().UTC().Unix()
	query := fmt.Sprintf("SELECT group_id, lesson_type, subject, subgroup_number, date, time FROM %s ORDER BY date-%[2]s WHERE date-%[2]s > 0 AND subject=%s AND group_id = %s LIMIT 4", LESSONS_TABLE, fmt.Sprint(date), subject, fmt.Sprint(groupId))
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 4)
	i := 0
	var storedDate, storedTime = "", ""
	for rows.Next() {
		err = rows.Scan(&lessons[i].GroupId, &lessons[i].LessonType, &lessons[i].Subject, &lessons[i].SubgroupNumber, &storedDate, &storedTime)
		if err != nil {
			return nil, err
		}
		lessons[i].Date, err = time.Parse(savedFormat, storedDate)
		if err != nil {
			return nil, err
		}
		lessons[i].Time, err = time.Parse(savedFormat, storedTime)
		if err != nil {
			return nil, err
		}
	}
	return lessons, nil
}

func (repo *LessonsRepository) GetSubjects(groupId int64) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT subject FROM %s WHERE group_id=$1 ORDER BY subject", LESSONS_TABLE)
	rows, err := repo.db.Query(query, groupId)
	if err != nil {
		return nil, err
	}
	subjects := []string{}
	var subject string
	for rows.Next() {
		err := rows.Scan(&subject)
		if err != nil {
			return nil, err
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		subjects = append(subjects, subject)
	}
	if len(subjects) == 0 {
		return nil, nil
	}
	return subjects, nil
}

func (repo *LessonsRepository) getSortedLessons(lessons []*entities.Lesson) []persistance.Lesson {
	if len(lessons) == 0 {
		return nil
	}
	storedLessons := make([]persistance.Lesson, 0, len(lessons)*3)
	filter := datastructures.NewOptimalBloomFiltet(len(lessons), 0.01)
	for _, lesson := range lessons {
		if lesson.LessonType != entities.Labwork {
			continue
		}
		checkedName := createCheckedName(lesson)
		exists := filter.Check(checkedName)
		if exists {
			notFalsePositive := slices.ContainsFunc(lessons, areLessonsEqual(lesson))
			if notFalsePositive {
				continue
			}
		}
		filter.Add(checkedName)

		startDate, endDate := time.Time(lesson.StartDate), time.Time(lesson.EndDate)
		currentDate := startDate
		for !currentDate.Equal(endDate) {
			addedTime := time.Hour * 24 * 7 * time.Duration(utils.CalculateWeeksDistance(lesson.WeekNumber, utils.CalculateWeek(startDate)))
			currentDate = currentDate.Add(addedTime)
			storedLesson := *persistance.FromLessonEntity(lesson, currentDate)
			storedLessons = append(storedLessons, storedLesson)
		}
	}

	sort.Slice(storedLessons, func(i, j int) bool {
		if !storedLessons[i].Date.Equal(storedLessons[j].Date) {
			return storedLessons[i].Date.Before(storedLessons[j].Date)
		}
		return storedLessons[i].Time.Before(storedLessons[j].Time)
	})

	return storedLessons
}

func createCheckedName(storedLesson *entities.Lesson) string {
	return storedLesson.Subject + storedLesson.StartDate.Format(time.DateOnly) + fmt.Sprint(storedLesson.SubgroupNumber)
}

func areLessonsEqual(self *entities.Lesson) func(other *entities.Lesson) bool {
	return func(other *entities.Lesson) bool {
		return time.Time(self.StartDate).Equal(time.Time(other.StartDate)) && self.SubgroupNumber == other.SubgroupNumber && self.Subject == other.Subject
	}
}
