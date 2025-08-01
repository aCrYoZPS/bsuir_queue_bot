package group

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
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
	bot    *tgbotapi.BotAPI
	groups GroupsRepository
	users  UsersRepository
}

func NewGroupSubmitState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groups GroupsRepository, users UsersRepository) *groupSubmitStartState {
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
	if user.Id == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже член группы")
		_, err := state.bot.Send(msg)
		return err
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_GROUPNAME_STATE))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите номер группы,в которую хотите вступить")
	_, err = state.bot.Send(msg)
	if err != nil {
		return err
	}
	return err
}

type groupSubmitGroupNameState struct {
	cache  interfaces.HandlersCache
	bot    *tgbotapi.BotAPI
	groups GroupsRepository
}

func NewGroupSubmitGroupNameState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groups GroupsRepository) *groupSubmitGroupNameState {
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
		_, err := state.bot.Send(msg)
		return err
	}
	form, err := json.Marshal(&groupSubmitForm{UserId: message.From.ID, UserName: message.From.UserName, Group: groupName})
	if err != nil {
		return err
	}
	state.cache.SaveInfo(ctx, message.Chat.ID, string(form))
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_SUBMIT_NAME_STATE))
	if err != nil {
		return err
	}
	_, err = state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введите ваши фамилию и имя (Пример формата: Иван Иванов)"))
	return err
}

type groupSubmitNameState struct {
	cache    interfaces.HandlersCache
	bot      *tgbotapi.BotAPI
	groups   GroupsRepository
	requests interfaces.RequestsRepository
}

func NewGroupSubmitNameState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groups GroupsRepository, requests interfaces.RequestsRepository) *groupSubmitNameState {
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
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введите фамилию и имя как в предоставленном образце"))
		return err
	}

	info, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return err
	}
	form := &groupSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}
	form.Name = fullName[0] + " " + fullName[1]
	data, err := json.Marshal(form)
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(data))
	if err != nil {
		return err
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.GROUP_WAITING_STATE))
	if err != nil {
		return err
	}

	admins, err := state.groups.GetAdmins(ctx, form.Group)
	if err != nil {
		return err
	}

	err = state.SendMessagesToAdmins(ctx, admins, form)
	if err != nil {
		return nil
	}
	_, err = state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ваша заявка была отправлена администратору"))
	return err
}

func (state *groupSubmitNameState) SendMessagesToAdmins(ctx context.Context, admins []entities.User, form *groupSubmitForm) error {
	errRes := make(chan error, 1)
	go func(chan error) {
		text := fmt.Sprintf("Пользователь под id @%s и именем \"%s\" хочет присоединиться к группе", form.UserName, form.Name)
		reqUUID := uuid.NewString()
		for _, admin := range admins {
			msg := tgbotapi.NewMessage(admin.TgId, text)
			msg.ReplyMarkup = createMarkupKeyboard(form)
			sentMsg, err := state.bot.Send(msg)
			if err != nil {
				errRes <- err
			}
			err = state.requests.SaveRequest(ctx, interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
			errRes <- err
		}
	}(errRes)
	var err error
	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to send messages to group admins: %v", ctx.Err())
	case err = <-errRes:
		return err
	}
}

type groupWaitingState struct {
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func NewGroupWaitingState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *groupWaitingState {
	return &groupWaitingState{cache: cache, bot: bot}
}

func (*groupWaitingState) StateName() string {
	return constants.GROUP_WAITING_STATE
}

func (state *groupWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	errRes := make(chan error, 1)
	go func(chan error) {
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ваша заявка всё ещё рассматривается, подождите"))
		errRes <- err
	}(errRes)
	var err error
	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to send waiting message: %v", ctx.Err())
	case err = <-errRes:
	    return err	
	}
}

func createMarkupKeyboard(form *groupSubmitForm) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.GROUP_CALLBACKS + "accept" + fmt.Sprint(form.UserId)
	declineData := constants.GROUP_CALLBACKS + "decline" + fmt.Sprint(form.UserId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Принять", acceptData), tgbotapi.NewInlineKeyboardButtonData("Отклонить", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
