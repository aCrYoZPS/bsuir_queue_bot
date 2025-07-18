package stateMachine

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminCallbackHandler struct {
	CallbackHandler
	repo  interfaces.UsersRepository
	cache interfaces.HandlersCache
}

func NewAdminCallbackHandler(repo interfaces.UsersRepository, cache interfaces.HandlersCache) *AdminCallbackHandler {
	return &AdminCallbackHandler{
		repo:  repo,
		cache: cache,
	}
}

func (handler *AdminCallbackHandler) HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return err
	}

	if !strings.HasPrefix(update.CallbackQuery.Data, ADMIN_CALLBACKS) {
		return errors.New("invalid command requested")
	}
	command := strings.TrimPrefix(update.CallbackQuery.Data, ADMIN_CALLBACKS)
	switch {
	case strings.HasPrefix(command, "accept"):
		var chatId int64
		chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
		if err != nil {
			return err
		}
		info, err := handler.cache.GetInfo(chatId)
		if err != nil {
			return err
		}
		form := &adminSubmitForm{}
		err = json.Unmarshal([]byte(info), form)
		if err != nil {
			return err
		}
		err = handler.repo.Add(entities.NewUser(form.Name, form.Group, chatId))
		if err != nil {
			return err
		}

		msg := tgbotapi.NewMessage(form.ChatId, "Ваша заявка была одобрена. Ссылка на гугл-таблицу: *ЗАГЛУШКА*")
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	case strings.HasPrefix(command, "decline"):
		var chatId int64
		err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
		if err != nil {
			return err
		}
		info, err := handler.cache.GetInfo(chatId)
		if err != nil {
			return err
		}
		form := &adminSubmitForm{}
		err = json.Unmarshal([]byte(info), &form)
		if err != nil {
			return err
		}
		msg := tgbotapi.NewMessage(form.ChatId, "Ваша заявка была отклонена. Причина: *ЗАГЛУШКА*")
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	default:
		return errors.New("no such callback")
	}
	return nil
}
