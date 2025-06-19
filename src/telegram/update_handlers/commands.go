package update_handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	HELP_COMMAND   = "help"
	SUBMIT_COMMAND = "submit"
)

var commands = []tgbotapi.BotCommand{
	{Command: HELP_COMMAND, Description: "Команды и информация"},
	{Command: SUBMIT_COMMAND, Description: "Yeah, I am lazy even for that"},
}

// Probably is not worthy enough to be a separate package at all, but for the readability, why not
func InitCommands() {
	tgbotapi.NewSetMyCommands(
		commands...,
	)
}
