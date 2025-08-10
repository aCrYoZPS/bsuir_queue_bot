package customlabworks

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlersCache interface {
	SaveState(context.Context, interfaces.CachedInfo) error
	GetState(ctx context.Context, chatId int64) (*interfaces.CachedInfo, error)
	SaveInfo(ctx context.Context, chatId int64, json string) error
	GetInfo(ctx context.Context, chatId int64) (string, error)
}

type LabworksService interface {
	AddLabwork(ctx context.Context, lesson *persistance.Lesson)
}

type UsersService interface {
	GetByTgId(ctx context.Context, tgId int64) (*entities.User, error)
}

type labworkAddStartState struct {
	bot   *tgutils.Bot
	cache HandlersCache
	users UsersService
}

func NewLabworkAddStartState(bot *tgutils.Bot, cache HandlersCache, users UsersService) *labworkAddStartState {
	return &labworkAddStartState{bot: bot, cache: cache, users: users}
}

func (*labworkAddStartState) StateName() string {
	return constants.LABWORK_ADD_START_STATE
}

func (state *labworkAddStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id in labwork add state: %w", err)
	}
	if !slices.Contains(user.Roles, entities.Admin) {
		err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition to idle state in labwork add start state: %w", err)
		}
		return nil
	}

	req := &LessonRequest{
		GroupId: user.GroupId,
	}
	jsonedRequest, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal lesson request during custom labwork add start state handling: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedRequest))
	if err != nil {
		return fmt.Errorf("failed to save jsoned request into cache during custom labwork add start state handling: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_ADD_SUBMIT_NAME_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to custom labwork submit name state: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите название добавленной пары"))
	if err != nil {
		return fmt.Errorf("failed to send response in labwork add start state: %w", err)
	}
	return nil
}

type labworkAddSubmitNameState struct {
	bot   *tgutils.Bot
	cache HandlersCache
}

func NewLabworkAddSubmitNameState(bot *tgutils.Bot, cache HandlersCache) *labworkAddSubmitNameState {
	return &labworkAddSubmitNameState{
		bot:   bot,
		cache: cache,
	}
}

func (*labworkAddSubmitNameState) StateName() string {
	return constants.LABWORK_ADD_SUBMIT_NAME_STATE
}

type LessonRequest struct {
	DateTime time.Time
	Name     string
	GroupId  int64
}

func (state *labworkAddSubmitNameState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	name := message.Text
	if name == "" {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Отправьте текстовое сообщение, без прикрепленных файлов"))
		if err != nil {
			return fmt.Errorf("failed to send no text response during custom labwork name submit: %w", err)
		}
		return nil
	}

	jsonedReq, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned request info in custom labwork name submit state: %w", err)
	}
	req := &LessonRequest{}
	err = json.Unmarshal([]byte(jsonedReq), &req)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned req (%s) in custom labwork name submit state: %w", jsonedReq, err)
	}
	req.Name = name
	json, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal lesson request in custom labwork submit name state: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(json))
	if err != nil {
		return fmt.Errorf("failed to save json string into cache in custom labwork submit name state %w", err)
	}

	resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите дату добавленной пары")
	resp.ReplyMarkup = createCalendar(time.Now(), true)

	_, err = state.bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send response during custom labwork name submit: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_ADD_WAITING_STATE))
	if err != nil {
		return err
	}

	return nil
}

type LabworkAddWaitingState struct {
	bot *tgutils.Bot
}

func NewLabworkAddWaitingState(bot *tgutils.Bot) *LabworkAddWaitingState {
	return &LabworkAddWaitingState{bot: bot}
}

func (*LabworkAddWaitingState) StateName() string {
	return constants.LABWORK_ADD_WAITING_STATE
}

func (state *LabworkAddWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Выберите дату и время пары"))
	if err != nil {
		return fmt.Errorf("failed to send message during labwork add waiting state: %w", err)
	}
	return nil
}
