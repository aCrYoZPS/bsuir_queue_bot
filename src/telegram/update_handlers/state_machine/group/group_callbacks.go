package group

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
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

func (handler *GroupCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return fmt.Errorf("failed to request callback from telegram: %w", err)
	}

	if !strings.HasPrefix(update.CallbackQuery.Data, constants.GROUP_CALLBACKS) {
		return fmt.Errorf("callback doesn't have group prefix (%s)", constants.GROUP_CALLBACKS)
	}
	command := strings.TrimPrefix(update.CallbackQuery.Data, constants.GROUP_CALLBACKS)
	var err error
	switch {
	case strings.HasPrefix(command, "accept"):
		err = handler.handleAcceptCallback(ctx, update.CallbackQuery.Message, command, bot)
	case strings.HasPrefix(command, "decline"):
		err = handler.handleDeclineCallback(ctx, update.CallbackQuery.Message, command, bot)
	default:
		err = fmt.Errorf("no handler from group callbacks to given command (%s)", command)
	}
	return err
}

func (handler *GroupCallbackHandler) handleAcceptCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return fmt.Errorf("failed to get info for group accept callback: %w", err)
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info in group accept callback: %w", err)
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state in group accept callback: %w", err)
	}

	err = handler.users.Add(ctx, entities.NewUser(form.Name, form.Group, form.UserId))
	if err != nil {
		return fmt.Errorf("failed to add new user in group accept callback: %w", err)
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return err
	}
	resp := tgbotapi.NewMessage(form.UserId, "Ваша заявка была одобрена")
	_, err = bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send response in group accept callback: %w", err)
	}
	return err
}

func (handler *GroupCallbackHandler) handleDeclineCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return fmt.Errorf("failed to unmarshal chat id from (%s) in group decline callback: %w", strings.TrimPrefix(command, "decline"), err)
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return fmt.Errorf("failed to get info in group decline callback: %w", err)
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal form (%s) in group decline callback: %w", info, err)
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to idle state in group decline callback")
	}

	resp := tgbotapi.NewMessage(form.UserId, "Ваша заявка была отклонена")
	if _, err := bot.SendCtx(ctx, resp); err != nil {
		return fmt.Errorf("failed to send response in group decline callback: %w", err)
	}
	return handler.RemoveMarkup(ctx, msg, bot)
}

func (handler *GroupCallbackHandler) RemoveMarkup(ctx context.Context, msg *tgbotapi.Message, bot *tgutils.Bot) error {
	request, err := handler.requests.GetByMsg(ctx, int64(msg.MessageID), msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get group request when removing markup: %w", err)
	}
	requests, err := handler.requests.GetByUUID(ctx, request.UUID)
	if err != nil {
		return fmt.Errorf("failed to get group request by uuid when removing markup: %w", err)
	}
	for _, request := range requests {
		err = handler.requests.DeleteRequest(ctx, request.MsgId)
		if err != nil {
			return fmt.Errorf("failed to delete group request during markup removal: %w", err)
		}
		_, err := bot.SendCtx(ctx,tgbotapi.NewEditMessageReplyMarkup(request.ChatId, int(request.MsgId), tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return fmt.Errorf("failed to send markup removal message: %w", err)
		}
	}
	return nil
}
