package group

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GroupsRepository interface {
	GetAdmins(groupName string) ([]entities.User, error)
	DoesGroupExist(groupName string) (bool, error)
}

type groupSubmitForm struct {
	UserId   int64  `json:"userId,omitempty"`
	UserName string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Group    string `json:"group,omitempty"`
}
type UsersRepository interface {
	GetByTgId(tgId int64) (*entities.User, error)
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

func (state *groupSubmitStartState) Handle(chatId int64, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(message.From.ID)
	if err != nil {
		return err
	}
	if user.Id == 0 {
		msg := tgbotapi.NewMessage(chatId, "Вы уже член группы")
		_, err := state.bot.Send(msg)
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.GROUP_SUBMIT_GROUPNAME_STATE))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "Введите номер группы,в которую хотите вступить")
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

func (state *groupSubmitGroupNameState) Handle(chatId int64, message *tgbotapi.Message) error {
	groupName := message.Text
	groupExists, err := state.groups.DoesGroupExist(groupName)
	if err != nil {
		return err
	}
	if !groupExists {
		msg := tgbotapi.NewMessage(chatId, "Данная группа не найдена")
		_, err := state.bot.Send(msg)
		return err
	}
	form, err := json.Marshal(&groupSubmitForm{UserId: message.From.ID, UserName: message.From.UserName, Group: groupName})
	if err != nil {
		return err
	}
	state.cache.SaveInfo(chatId, string(form))
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.GROUP_SUBMIT_NAME_STATE))
	if err != nil {
		return err
	}
	_, err = state.bot.Send(tgbotapi.NewMessage(chatId, "Введите ваши фамилию и имя (Пример формата: Иван Иванов)"))
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

func (state *groupSubmitNameState) Handle(chatId int64, message *tgbotapi.Message) error {
	name := message.Text
	fullName := strings.Fields(name)
	if len(fullName) != 2 {
		_, err := state.bot.Send(tgbotapi.NewMessage(chatId, "Введите фамилию и имя как в предоставленном образце"))
		return err
	}

	info, err := state.cache.GetInfo(chatId)
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
	err = state.cache.SaveInfo(chatId, string(data))
	if err != nil {
		return err
	}

	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	admins, err := state.groups.GetAdmins(form.Group)
	if err != nil {
		return err
	}
	err = state.SendMessagesToAdmins(admins, form)
	if err != nil {
		return nil
	}
	_, err = state.bot.Send(tgbotapi.NewMessage(chatId, "Ваша заявка была отправлена администратору"))
	return err
}

func (state *groupSubmitNameState) SendMessagesToAdmins(admins []entities.User, form *groupSubmitForm) error {
	text := fmt.Sprintf("Пользователь под id @%s и именем \"%s\" хочет присоединиться к группе", form.UserName, form.Name)
	for _, admin := range admins {
		msg := tgbotapi.NewMessage(admin.TgId, text)
		msg.ReplyMarkup = createMarkupKeyboard(form)
		sentMsg, err := state.bot.Send(msg)
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(interfaces.NewRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID))
		if err != nil {
			return err
		}
	}
	return nil
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

func (state *groupWaitingState) Handle(chatId int64, message *tgbotapi.Message) error {
	_, err := state.bot.Send(tgbotapi.NewMessage(chatId, "Ваша заявка всё ещё рассматривается, подождите"))
	return err
}

func createMarkupKeyboard(form *groupSubmitForm) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.GROUP_CALLBACKS + "accept" + fmt.Sprint(form.UserId)
	declineData := constants.GROUP_CALLBACKS + "decline" + fmt.Sprint(form.UserId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Принять", acceptData), tgbotapi.NewInlineKeyboardButtonData("Отклонить", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
