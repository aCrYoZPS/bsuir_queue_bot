package ioc

import (
	"database/sql"
	"log/slog"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces/mocks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

var useSqliteConnection = provider(
	func() *sql.DB {
		conn, err := sql.Open("sqlite3", "./sqlite3.db")
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		return conn
	},
)

var useMockGroupsRepository = provider(
	func() interfaces.GroupsRepository {
		return mocks.NewGroupsRepositoryMock()
	},
)
var useGroupsGepository = provider(
	func() interfaces.GroupsRepository {
		return sqlite.NewGroupsRepository(
			useSqliteConnection(),
		)
	},
)

var useMockLessonsRepository = provider(
	func() interfaces.LessonsRepository {
		return mocks.NewLessonsRepositoryMock()
	},
)
var useLessonsRepository = provider(
	func() interfaces.LessonsRepository {
		return sqlite.NewLessonsRepository(
			useSqliteConnection(),
		)
	},
)
