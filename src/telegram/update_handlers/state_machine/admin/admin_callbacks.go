package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminCallbackHandler struct {
	usersRepo interfaces.UsersRepository
	sheets    sheetsapi.SheetsApi
	cache     interfaces.HandlersCache
}

func NewAdminCallbackHandler(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache, sheets sheetsapi.SheetsApi) *AdminCallbackHandler {
	return &AdminCallbackHandler{
		usersRepo: usersRepo,
		cache:     cache,
		sheets:    sheets,
	}
}

func (handler *AdminCallbackHandler) HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return err
	}

	if !strings.HasPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS) {
		return errors.New("invalid command requested")
	}
	command := strings.TrimPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS)
	switch {
	case strings.HasPrefix(command, "accept"):
		handler.handleAcceptCallback(command, bot)
	case strings.HasPrefix(command, "decline"):
		handler.handleDeclineCallback(command, bot)
	default:
		return errors.New("no such callback")
	}
	return nil
}

func (handler *AdminCallbackHandler) handleAcceptCallback(command string, bot *tgbotapi.BotAPI) error {
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

	err = handler.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	err = handler.usersRepo.Add(entities.NewUser(form.Name, form.Group, form.UserId))
	if err != nil {
		return err
	}

	url, err := handler.sheets.CreateSheet(form.Group)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(form.UserId, fmt.Sprintf("Ваша заявка была одобрена. Ссылка на гугл-таблицу: %s", url))
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (handler *AdminCallbackHandler) handleDeclineCallback(command string, bot *tgbotapi.BotAPI) error {
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
	msg := tgbotapi.NewMessage(form.UserId, "Ваша заявка была отклонена")
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}
