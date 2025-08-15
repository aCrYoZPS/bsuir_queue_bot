package tgutils

import (
	"context"
	"fmt"
	"slices"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	JOIN_GROUP_KEYBOARD  = "Вступить в группу"
	ASSIGN_KEYBOARD      = "Стать администратором группы"
	SUBMIT_KEYBOARD      = "Отправить заявку на лабораторную"
	ADD_LABWORK_KEYBOARD = "Добавить собственную лабораторную"
)

func CreateStartReplyMarkup(ctx context.Context, msg *tgbotapi.MessageConfig, user *entities.User, bot *Bot) error {
	keyboard := []tgbotapi.KeyboardButton{}
	if user.GroupId == 0 {
		keyboard = append(keyboard, tgbotapi.KeyboardButton{Text: JOIN_GROUP_KEYBOARD})
	}
	if !slices.Contains(user.Roles, entities.Admin) {
		keyboard = append(keyboard, tgbotapi.KeyboardButton{Text: ASSIGN_KEYBOARD})
	}
	if user.GroupId != 0 {
		keyboard = append(keyboard, tgbotapi.KeyboardButton{Text: SUBMIT_KEYBOARD})
	}
	if slices.Contains(user.Roles, entities.Admin) {
		keyboard = append(keyboard, tgbotapi.KeyboardButton{Text: ADD_LABWORK_KEYBOARD})
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
	_, err := bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to create start reply markup: %w", err)
	}
	return nil
}
