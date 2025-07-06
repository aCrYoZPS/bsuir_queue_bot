package ioc

import (
	"database/sql"
	"log/slog"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces/mocks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/memory"
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
		err = sqlite.DatabaseInit(conn)
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

var useHandlersCache = provider(
	func() interfaces.HandlersCache {
		repo := memory.NewHandlersCache()
		return repo
	},
)

var useGroupsGepository = provider(
	func() interfaces.GroupsRepository {
		repo, err := sqlite.NewGroupsRepository(
			useSqliteConnection(),
		)
		if err != nil {
			slog.Error(err.Error())
			panic(err)
		}
		return repo
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
