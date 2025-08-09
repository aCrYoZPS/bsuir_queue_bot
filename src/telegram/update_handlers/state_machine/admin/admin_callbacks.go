package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	adminInterfaces "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/admin/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var ErrAlreadyAdmin = errors.New("the user is already admin to the group")

type AdminCallbackHandler struct {
	usersRepo interfaces.UsersRepository
	requests  interfaces.AdminRequestsRepository
	cache     interfaces.HandlersCache
	lessons   adminInterfaces.LessonsService
}

func NewAdminCallbackHandler(usersRepo interfaces.UsersRepository, cache interfaces.HandlersCache, requests interfaces.AdminRequestsRepository, lessons adminInterfaces.LessonsService) *AdminCallbackHandler {
	return &AdminCallbackHandler{
		usersRepo: usersRepo,
		cache:     cache,
		requests:  requests,
		lessons:   lessons,
	}
}

func (handler *AdminCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	if !strings.HasPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS) {
		return errors.New("invalid command requested")
	}
	command := strings.TrimPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS)
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

func (handler *AdminCallbackHandler) handleAcceptCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	err = handler.addAdmin(ctx, form)
	if err != nil {
		return err
	}

	url, err := handler.lessons.AddGroupLessons(ctx, form.Group)
	if err != nil {
		return err
	}

	resp := tgbotapi.NewMessage(form.UserId, fmt.Sprintf("Ваша заявка была одобрена. Ссылка на гугл-таблицу: %s", url))
	if _, err := bot.Send(resp); err != nil {
		return err
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	return err
}

func (handler *AdminCallbackHandler) addAdmin(ctx context.Context,form *adminSubmitForm) error {
	user, err := handler.usersRepo.GetByTgId(ctx, form.UserId)
	if err != nil {
		return err
	}
	if slices.Contains(user.Roles, entities.Admin) {
		return ErrAlreadyAdmin
	}
	user.Roles = append(user.Roles, entities.Admin)
	err = handler.usersRepo.Update(ctx, user)
	return err
}

func (handler *AdminCallbackHandler) handleDeclineCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return err
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return err
	}
	resp := tgbotapi.NewMessage(form.UserId, "Ваша заявка была отклонена")
	_, err = bot.Send(resp)
	return err
}

func (handler *AdminCallbackHandler) RemoveMarkup(ctx context.Context, msg *tgbotapi.Message, bot *tgutils.Bot) error {
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
