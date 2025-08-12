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

func NewLessonsRepository(db *sql.DB) *LessonsRepository {
	repo := &LessonsRepository{
		db: db,
	}
	return repo
}

const savedFormat = time.RFC3339

func (repo *LessonsRepository) AddRange(ctx context.Context, lessons []*entities.Lesson) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}
	storedLessons := repo.getSortedLessons(ctx, lessons)
	for _, lesson := range storedLessons {
		storedDate := lesson.Date.Format(savedFormat)
		storedTime := lesson.Time.Format(savedFormat)
		query := fmt.Sprintf("INSERT INTO %s (group_id, subject, lesson_type, subgroup_number, date, time) values ($1,$2,$3,$4,$5,$6)", LESSONS_TABLE)
		tx.ExecContext(ctx, query, lesson.GroupId, lesson.Subject, lesson.LessonType, lesson.SubgroupNumber, storedDate, storedTime)
	}
	err = tx.Commit()
	return err
}

func (repo *LessonsRepository) Add(ctx context.Context, lesson *persistance.Lesson) error {
	query := fmt.Sprintf("INSERT INTO %s (group_id, subject, lesson_type, subgroup_number, date, time) values ($1,$2,$3,$4,$5,$6)", LESSONS_TABLE)
	storedDate := lesson.Date.Format(savedFormat)
	storedTime := lesson.Time.Format(savedFormat)
	_, err := repo.db.ExecContext(ctx, query, lesson.GroupId, lesson.Subject, lesson.LessonType, lesson.SubgroupNumber, storedDate, storedTime)
	return err
}

func (repo *LessonsRepository) GetAll(ctx context.Context, groupName string) ([]persistance.Lesson, error) {
	query := fmt.Sprintf("SELECT id, group_id, lesson_type, subject, subgroup_number, date, time FROM %s WHERE $1 in (SELECT name from %s)", LESSONS_TABLE, GROUPS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 0, 100)
	i := 0
	var storedTime, storedDate = "", ""
	for rows.Next() {
		lesson := &persistance.Lesson{}
		err = rows.Scan(&lesson.Id, &lesson.GroupId, &lesson.LessonType, &lesson.Subject, &lesson.SubgroupNumber, &storedDate, &storedTime)
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

func (repo *LessonsRepository) GetNext(ctx context.Context, subject string, groupId int64) ([]persistance.Lesson, error) {
	date := time.Now().UTC().Unix()
	query := fmt.Sprintf("SELECT id, group_id, lesson_type, subject, subgroup_number, date, time FROM %s WHERE strftime('%%s', date)>$1 AND subject=$2 AND group_id = $3 ORDER BY date LIMIT 4", LESSONS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query, fmt.Sprint(date), subject, groupId)
	if err != nil {
		return nil, err
	}
	lessons := make([]persistance.Lesson, 4)
	i := 0
	var storedDate, storedTime = "", ""
	for rows.Next() {
		err = rows.Scan(&lessons[i].Id, &lessons[i].GroupId, &lessons[i].LessonType, &lessons[i].Subject, &lessons[i].SubgroupNumber, &storedDate, &storedTime)
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
		i++
	}
	if i == 0 {
		return []persistance.Lesson{}, nil
	}
	return lessons[:i], nil
}

func (repo *LessonsRepository) GetSubjects(ctx context.Context, groupId int64) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT subject FROM %s WHERE group_id=$1 AND date>=date('now') ORDER BY subject", LESSONS_TABLE)
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

func (repo *LessonsRepository) getSortedLessons(ctx context.Context, lessons []*entities.Lesson) []persistance.Lesson {
	resChan := make(chan []persistance.Lesson, 1)
	go func(chan []persistance.Lesson) {
		if len(lessons) == 0 {
			resChan <- nil
		}
		storedLessons := make([]persistance.Lesson, 0, len(lessons)*3)
		filter := datastructures.NewOptimalBloomFilter(len(lessons), 0.01)
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
		resChan <- storedLessons
	}(resChan)

	select {
	case res := <-resChan:
		return res
	case <-ctx.Done():
		return nil
	}
}

func createCheckedName(storedLesson *entities.Lesson) string {
	return storedLesson.Subject + storedLesson.StartDate.Format(time.DateOnly) + fmt.Sprint(storedLesson.SubgroupNumber)
}

func areLessonsEqual(self *entities.Lesson) func(other *entities.Lesson) bool {
	return func(other *entities.Lesson) bool {
		return time.Time(self.StartDate).Equal(time.Time(other.StartDate)) && self.SubgroupNumber == other.SubgroupNumber && self.Subject == other.Subject
	}
}

func (repo *LessonsRepository) GetEndedLessons(ctx context.Context, before time.Time) ([]persistance.Lesson, error) {
	lessons := []persistance.Lesson{}
	query := fmt.Sprintf("SELECT id, group_id, subject, lesson_type, subgroup_number, date, time FROM %s WHERE date >= date('now') ORDER BY date", LESSONS_TABLE)
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var (
		appendedLesson         persistance.Lesson
		storedTime, storedDate string
	)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		err := rows.Scan(appendedLesson.Id, appendedLesson.GroupId, appendedLesson.Subject, appendedLesson.SubgroupNumber, storedDate, storedTime)
		if err != nil {
			return nil, err
		}
		appendedLesson.Date, err = time.Parse(savedFormat, storedDate)
		if err != nil {
			return nil, err
		}
		appendedLesson.Time, err = time.Parse(savedFormat, storedTime)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, appendedLesson)
	}
	return lessons, nil
}

func (repo *LessonsRepository) GetLessonByRequest(ctx context.Context, requestId int64) (*persistance.Lesson, error) {
	query := fmt.Sprintf("SELECT l.id, l.group_id, l.lesson_type, l.subject, l.subgroup_number, l.date, l.time FROM %s AS l INNER JOIN %s as r ON r.lesson_id = l.id WHERE r.id=$1 ", LESSONS_TABLE, LESSONS_REQUESTS_TABLE)
	row := repo.db.QueryRowContext(ctx, query, requestId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var (
		lesson                 persistance.Lesson
		storedDate, storedTime string
	)
	err := row.Scan(&lesson.Id, &lesson.GroupId, &lesson.LessonType, &lesson.Subject, &lesson.SubgroupNumber, &storedDate, &storedTime)
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
	return &lesson, nil
}

func (repo *LessonsRepository) DeleteLessons(ctx context.Context, before time.Time) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE date-$1 < 0", LESSONS_TABLE)
	_, err := repo.db.ExecContext(ctx, query, before)
	return err
}
