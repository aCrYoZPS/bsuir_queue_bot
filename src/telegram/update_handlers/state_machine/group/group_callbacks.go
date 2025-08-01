package group

import (
	"context"
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
	users    interfaces.UsersRepository
	requests interfaces.RequestsRepository
	cache    interfaces.HandlersCache
}

func NewGroupCallbackHandler(users interfaces.UsersRepository, cache interfaces.HandlersCache, requests interfaces.RequestsRepository) *GroupCallbackHandler {
	return &GroupCallbackHandler{
		users:    users,
		cache:    cache,
		requests: requests,
	}
}

func (handler *GroupCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
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
		err = handler.handleAcceptCallback(ctx, update.CallbackQuery.Message, command, bot)
	case strings.HasPrefix(command, "decline"):
		err = handler.handleDeclineCallback(ctx, update.CallbackQuery.Message, command, bot)
	default:
		err = errors.New("no such callback")
	}
	return err
}

func (handler *GroupCallbackHandler) handleAcceptCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	err = handler.users.Add(ctx, entities.NewUser(form.Name, form.Group, form.UserId))
	if err != nil {
		return err
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return err
	}
	resp := tgbotapi.NewMessage(form.UserId, "Ваша заявка была одобрена")
	_, err = bot.Send(resp)
	return err
}

func (handler *GroupCallbackHandler) handleDeclineCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return err
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	resp := tgbotapi.NewMessage(form.UserId, "Ваша заявка была отклонена")
	if _, err := bot.Send(resp); err != nil {
		return err
	}
	return handler.RemoveMarkup(ctx, msg, bot)
}

func (handler *GroupCallbackHandler) RemoveMarkup(ctx context.Context, msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	request, err := handler.requests.GetByMsg(ctx, int64(msg.MessageID), msg.Chat.ID)
	if err != nil {
		return err
	}
	requests, err := handler.requests.GetByUUID(ctx, request.UUID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		err = handler.requests.DeleteRequest(ctx, request.MsgId)
		if err != nil {
			return err
		}
		_, err := bot.Send(tgbotapi.NewEditMessageReplyMarkup(request.ChatId, int(request.MsgId), tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return err
		}
	}
	return nil
}
