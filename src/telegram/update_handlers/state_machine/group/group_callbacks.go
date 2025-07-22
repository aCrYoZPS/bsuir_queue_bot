package group

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GroupCallbackHandler struct {
	usersRepo interfaces.UsersRepository
	cache     interfaces.HandlersCache
}

func NewGroupCallbackHandler(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache) *GroupCallbackHandler {
	return &GroupCallbackHandler{
		usersRepo: usersRepo,
		cache:     cache,
	}
}

func (handler *GroupCallbackHandler) HandleCallback(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return err
	}

	if !strings.HasPrefix(update.CallbackQuery.Data, constants.GROUP_CALLBACKS) {
		return errors.New("invalid command requested")
	}
	command := strings.TrimPrefix(update.CallbackQuery.Data, constants.GROUP_CALLBACKS)
	var err error
	switch {
	case strings.HasPrefix(command, "accept"):
		err = handler.handleAcceptCallback(command, bot)
	case strings.HasPrefix(command, "decline"):
		err = handler.handleDeclineCallback(command, bot)
	default:
		err = errors.New("no such callback")
	}
	return err
}

func (handler *GroupCallbackHandler) handleAcceptCallback(command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(chatId)
	if err != nil {
		return err
	}
	form := &groupSubmitForm{}
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

	msg := tgbotapi.NewMessage(form.UserId, "Ваша заявка была одобрена")
	_, err = bot.Send(msg)
	return err
}

func (handler *GroupCallbackHandler) handleDeclineCallback(command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(chatId)
	if err != nil {
		return err
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return err
	}

	err = handler.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(form.UserId, "Ваша заявка была отклонена")
	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}
