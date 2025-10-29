package ioc

import (
	"database/sql"
	"os"

	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces/mocks"
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

var useRequestsRepository = provider(
	func() *sqlite.RequestsRepository {
		return sqlite.NewRequestsRepository(useSqliteConnection())
	},
)

var useAdminRequestsRepository = provider(
	func() *sqlite.AdminRequestsRepository {
		return sqlite.NewAdminRequestsRepository(useSqliteConnection())
	},
)

var useHandlersCache = provider(
	func() *sqlite.HandlersCache {
		repo := sqlite.NewHandlersCache(useSqliteConnection())
		return repo
	},
)

var useGroupsRepository = provider(
	func() *sqlite.GroupsRepository {
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

var useUsersRepository = provider(
	func() *sqlite.UsersRepository {
		return sqlite.NewUsersRepository(useSqliteConnection())
	},
)

var useLessonsRepository = provider(
	func() *sqlite.LessonsRepository {
		repos := sqlite.NewLessonsRepository(
			useSqliteConnection(),
		)
		return repos
	},
)

var useLessonsRequestsRepository = provider(
	func() *sqlite.LessonsRequestsRepository {
		repos := sqlite.NewLessonsRequestsRepository(
			useSqliteConnection(),
		)
		return repos
	},
)

var useTasksRepository = provider(
	func() *sqlite.TasksRepository {
		return sqlite.NewTasksRepository(useSqliteConnection())
	},
)
