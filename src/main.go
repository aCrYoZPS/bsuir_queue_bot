package main

import (
	"database/sql"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/ioc"
	logging "github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite"
	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"

	"github.com/joho/godotenv"
)

func main() {
	logging.InitLogging()

	err := godotenv.Load()
	if err != nil {
		logging.FatalLog("Error loading .env file")
	}

	db, _ := sql.Open("sqlite3", "sqlite3.db")
	var gr interfaces.GroupsRepository
	gr = sqlite.NewGroupsRepository(db)
	gr.Init()

	err = gr.DeleteGroup(1)
	if err != nil {
		logging.FatalLog("failed to delete group", "err", err)
	}

	// TODO: CHANGE TO STARTUP SCRIPT OF SOME KIND
	srv := ioc.UseSheetsApiService()
	err = srv.CreateSheets()
	if err != nil {
		logging.Error("failed to init sheets", "err", err.Error())
	}

	bot.InitBot()
}
