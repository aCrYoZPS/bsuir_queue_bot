package ioc

import (
	"database/sql"
	"os"

	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces/mocks"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/memory"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

var useSqliteConnection = provider(
	func() *sql.DB {
		conn, err := sql.Open("sqlite3", os.Getenv("SQLITE_FILE"))
		if err != nil {
			logging.FatalLog(err.Error())
		}
		err = sqlite.DatabaseInit(conn)
		if err != nil {
			logging.FatalLog(err.Error())
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
		repos, err := sqlite.NewGroupsRepository(
			useSqliteConnection(),
		)
		if err != nil {
			logging.FatalLog(err.Error())
		}

		return repos
	},
)

var useMockLessonsRepository = provider(
	func() interfaces.LessonsRepository {
		return mocks.NewLessonsRepositoryMock()
	},
)

var useLessonsRepository = provider(
	func() interfaces.LessonsRepository {
		repos := sqlite.NewLessonsRepository(
			useSqliteConnection(),
		)
		return repos
	},
)
