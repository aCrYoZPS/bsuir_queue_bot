package stateMachine

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StateName string

type StateMachine struct {
	update_handlers.StateMachine
	cache     interfaces.HandlersCache
	bot       *tgbotapi.BotAPI
	usersRepo interfaces.UsersRepository
}

type statesConfig struct {
	cache      interfaces.HandlersCache
	bot        *tgbotapi.BotAPI
	groupsRepo interfaces.GroupsRepository
	usersRepo  interfaces.UsersRepository
	requests   interfaces.RequestsRepository
}

func NewStatesConfig(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, groupsRepo interfaces.GroupsRepository, usersRepo interfaces.UsersRepository, requests interfaces.RequestsRepository) *statesConfig {
	return &statesConfig{
		cache:      cache,
		bot:        bot,
		groupsRepo: groupsRepo,
		usersRepo:  usersRepo,
		requests:   requests,
	}
}

func NewStateMachine(conf *statesConfig) *StateMachine {
	InitStates(conf)
	return &StateMachine{cache: conf.cache, bot: conf.bot, usersRepo: conf.usersRepo}
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
