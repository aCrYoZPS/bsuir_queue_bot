package group

import (
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GroupsRepository interface {
	GetAdmins(groupName string) ([]entities.User, error)
	DoesGroupExist(groupName string) (bool, error)
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
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.GROUP_SUBMIT_NAME_STATE))
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
	cache    interfaces.HandlersCache
	bot      *tgbotapi.BotAPI
	groups   GroupsRepository
	requests interfaces.RequestsRepository
}

func NewGroupSubmitNameState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groups GroupsRepository) *groupSubmitGroupNameState {
	return &groupSubmitGroupNameState{bot: bot, cache: cache, groups: groups}
}

func (*groupSubmitGroupNameState) StateName() string {
	return constants.GROUP_SUBMIT_NAME_STATE
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
	admins, err := state.groups.GetAdmins(groupName)
	if err != nil {
		return err
	}
	if admins == nil {
		msg := tgbotapi.NewMessage(chatId, "У данной группы пока нет одобренных администраторов")
		_, err := state.bot.Send(msg)
		return err
	}

	err = state.SendMessagesToAdmins(admins, message.From.UserName)
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}
	_, err = state.bot.Send(tgbotapi.NewMessage(chatId, "Ваша заявка была отправлена администратору группы"))
	return err
}

func (state *groupSubmitGroupNameState) SendMessagesToAdmins(admins []entities.User, fromUsername string) error {
	text := "Пользователь под именем @%s хочет присоединиться к группе"
	for _, admin := range admins {
		msg, err := state.bot.Send(tgbotapi.NewMessage(admin.TgId, fmt.Sprintf(text, fromUsername)))
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(interfaces.NewRequest(int64(msg.MessageID), msg.Chat.ID))
		if err != nil {
			return err
		}
	}
	return nil
}
