package group

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type GroupsRepository interface {
	GetAdmins(ctx context.Context, groupName string) ([]entities.User, error)
	DoesGroupExist(ctx context.Context, groupName string) (bool, error)
}

type groupSubmitForm struct {
	UserId   int64  `json:"userId,omitempty"`
	UserName string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Group    string `json:"group,omitempty"`
}
type UsersRepository interface {
	GetByTgId(ctx context.Context, tgId int64) (*entities.User, error)
}

type groupSubmitStartState struct {
	cache  interfaces.HandlersCache
	bot    *tgutils.Bot
	groups GroupsRepository
	users  UsersRepository
}

func NewGroupSubmitState(cache interfaces.HandlersCache, bot *tgutils.Bot, groups GroupsRepository, users UsersRepository) *groupSubmitStartState {
	return &groupSubmitStartState{cache: cache, bot: bot, groups: groups, users: users}
}

func (*groupSubmitStartState) StateName() string {
	return constants.GROUP_SUBMIT_START_STATE
}

func (state *groupSubmitStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return err
	}
	if user.GroupId != 0 {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Вы уже член группы"))
		if err != nil {
			return fmt.Errorf("failed to send part of group response during group submit start state: %w", err)
		}
		err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition to idle state during group submit start state: %w", err)
		}
		return nil
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_GROUPNAME_STATE))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите номер группы,в которую хотите вступить")
	_, err = state.bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send group number request when starting group submit: %w", err)
	}
	return err
}

type groupSubmitGroupNameState struct {
	cache  interfaces.HandlersCache
	bot    *tgutils.Bot
	groups GroupsRepository
}

func NewGroupSubmitGroupNameState(cache interfaces.HandlersCache, bot *tgutils.Bot, groups GroupsRepository) *groupSubmitGroupNameState {
	return &groupSubmitGroupNameState{bot: bot, cache: cache, groups: groups}
}

func (*groupSubmitGroupNameState) StateName() string {
	return constants.GROUP_SUBMIT_GROUPNAME_STATE
}

func (state *groupSubmitGroupNameState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	groupName := message.Text
	groupExists, err := state.groups.DoesGroupExist(ctx, groupName)
	if err != nil {
		return err
	}
	if !groupExists {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Данная группа не найдена")
		_, err := state.bot.SendCtx(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to send not found message during submitting group submit name state: %w", err)
		}
		return nil
	}

	admins, err := state.groups.GetAdmins(ctx, groupName)
	if err != nil {
		return fmt.Errorf("failed to get group admins during group submit groupname state: %w", err)
	}
	if len(admins) == 0 {
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
		if err != nil {
			return fmt.Errorf("failed to save idle state during group name submit: %w", err)
		}
		_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "У данной группы пока нет администраторов. Попросите кого-либо из участников группы выступить в его роли"))
		if err != nil {
			return fmt.Errorf("failed to send message during group name submit: %w", err)
		}
		return nil
	}

	form, err := json.Marshal(&groupSubmitForm{UserId: message.From.ID, UserName: message.From.UserName, Group: groupName})
	if err != nil {
		return fmt.Errorf("failed to marshal group submit form: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(form))
	if err != nil {
		return fmt.Errorf("failed to save group submit form %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_NAME_STATE))
	if err != nil {
		return fmt.Errorf("failed to save group submit name state: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите ваши фамилию и имя (Пример формата: Иван Иванов)"))
	if err != nil {
		return fmt.Errorf("failed to send message for submitting user info: %w", err)
	}
	return nil
}

type groupSubmitNameState struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	groups   GroupsRepository
	requests interfaces.RequestsRepository
}

func NewGroupSubmitNameState(cache interfaces.HandlersCache, bot *tgutils.Bot, groups GroupsRepository, requests interfaces.RequestsRepository) *groupSubmitNameState {
	return &groupSubmitNameState{
		cache:    cache,
		bot:      bot,
		groups:   groups,
		requests: requests,
	}
}

func (state *groupSubmitNameState) StateName() string {
	return constants.GROUP_SUBMIT_NAME_STATE
}

func (state *groupSubmitNameState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	name := message.Text
	fullName := strings.Fields(name)
	if len(fullName) != 2 {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите фамилию и имя как в предоставленном образце"))
		return fmt.Errorf("failed to send message when submiting name for attendance to group: %w", err)
	}

	info, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get group submit info: %w", err)
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info (%s) into group submit form: %w", string(info), err)
	}
	form.Name = fullName[0] + " " + fullName[1]
	data, err := json.Marshal(form)
	if err != nil {
		return fmt.Errorf("failed to marshal group submit form: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(data))
	if err != nil {
		return fmt.Errorf("failed to save info: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	admins, err := state.groups.GetAdmins(ctx, form.Group)
	if err != nil {
		return fmt.Errorf("failed to get group admins during group name submit: %w", err)
	}

	err = state.SendMessagesToAdmins(ctx, admins, form)
	if err != nil {
		return fmt.Errorf("failed to send messages to admins during group submit name state: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Ваша заявка была отправлена администраторам группы"))
	if err != nil {
		return fmt.Errorf("failed to send message during group name submit: %w", err)
	}
	return nil
}

func (state *groupSubmitNameState) SendMessagesToAdmins(ctx context.Context, admins []entities.User, form *groupSubmitForm) error {
	if len(admins) == 0 {
		return errors.New("no admins found in group")
	}
	text := fmt.Sprintf("Пользователь под id @%s и именем \"%s\" хочет присоединиться к группе", form.UserName, form.Name)
	reqUUID := uuid.NewString()
	for _, admin := range admins {
		msg := tgbotapi.NewMessage(admin.TgId, text)
		msg.ReplyMarkup = createMarkupKeyboard(form)
		sentMsg, err := state.bot.SendCtx(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to send messages to admin during submitting group name: %w", err)
		}
		err = state.requests.SaveRequest(ctx, interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
		if err != nil {
			return fmt.Errorf("failed to save group request while sending messages to admin: %w", err)
		}
	}
	return nil
}

type groupWaitingState struct {
	cache interfaces.HandlersCache
	bot   *tgutils.Bot
}

func NewGroupWaitingState(cache interfaces.HandlersCache, bot *tgutils.Bot) *groupWaitingState {
	return &groupWaitingState{cache: cache, bot: bot}
}

func (*groupWaitingState) StateName() string {
	return constants.GROUP_WAITING_STATE
}

func (state *groupWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Ваша заявка всё ещё рассматривается, подождите"))
	if err != nil {
		return fmt.Errorf("failed to send message to user waiting for group submit: %w", err)
	}
	return nil
}

func createMarkupKeyboard(form *groupSubmitForm) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.GROUP_CALLBACKS + "accept" + fmt.Sprint(form.UserId)
	declineData := constants.GROUP_CALLBACKS + "decline" + fmt.Sprint(form.UserId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Принять", acceptData), tgbotapi.NewInlineKeyboardButtonData("Отклонить", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
