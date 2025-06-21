package update_handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	HELP_COMMAND   = "help"
	SUBMIT_COMMAND = "submit"
	ASSIGN_COMMAND = "assign"
)

var commands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Yeah, I am lazy even for that"},
	{Command: ASSIGN_COMMAND, Description: "Submit your request for becoming admin of the chosen group"},
}

// Probably is not worthy enough to be a separate package at all, but for the readability, why not
func InitCommands() {
	tgbotapi.NewSetMyCommands(
		commands...,
	)
}
