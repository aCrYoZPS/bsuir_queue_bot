package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateName string

type GroupsService interface {
	DoesGroupExist(name string) (bool, error)
}

type StateMachine struct {
	update_handlers.StateMachine
	cache     interfaces.HandlersCache
	bot       *tgbotapi.BotAPI
	groupsSrv *GroupsService
}

type statesConfig struct {
	cache         interfaces.HandlersCache
	bot           *tgbotapi.BotAPI
	groupsService GroupsService
	usersRepo     interfaces.UsersRepository
}

func NewStatesConfig(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groupsService GroupsService, usersRepo interfaces.UsersRepository) *statesConfig {
	return &statesConfig{
		cache:         cache,
		bot:           bot,
		groupsService: groupsService,
		usersRepo:     usersRepo,
	}
}

func NewStateMachine(conf *statesConfig) *StateMachine {
	InitStates(conf)
	return &StateMachine{cache: conf.cache, bot: conf.bot, groupsSrv: &conf.groupsService}
}

func (machine *StateMachine) HandleState(chatId int64, message *tgbotapi.Message) error {
	mu := machine.cache.AcquireLock(chatId)
	mu.Lock()

	defer mu.Unlock()
	defer machine.cache.ReleaseLock(chatId)

	info, err := machine.cache.GetState(chatId)
	if err != nil {
		return err
	}

	var state State
	if info != nil {
		state, err = getStateByName(info.State())
		if err != nil {
			return err
		}
	} else {
		state, err = getStateByName(constants.IDLE_STATE)
		if err != nil {
			return err
		}
	}

	return state.Handle(chatId, message)
}
