package requestdelete

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RequestsRepository interface {
	Delete(ctx context.Context, id int64) error
}

type UsersRepository interface {
	GetByTgId(ctx context.Context, tgId int64) (entities.User, error)
	Get(ctx context.Context, id int64) (entities.User, error)
}

type DeleteStartState struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	requests RequestsRepository
	users    UsersRepository
	lessons  LessonsRepository
}

const SubjectsChunk = 3

func NewDeleteStartState(bot *tgutils.Bot, cache interfaces.HandlersCache, requests RequestsRepository, users UsersRepository, lessons LessonsRepository) *DeleteStartState {
	return &DeleteStartState{cache: cache, bot: bot, requests: requests, users: users, lessons: lessons}
}

type DeleteStatesInfo struct {
	SentMsgId int                      `json:"sent,omitempty"`
	Requests  []entities.LessonRequest `json:"requests,omitempty"`
}

func (state *DeleteStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	responseText := "Выберите предмет для удаления"
	user, err := state.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get current user in request delete start state: %w", err)
	}
	subjects, err := state.lessons.GetSubjects(ctx, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get subjects for the group in request delete start state: %w", err)
	}
	markup := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, len(subjects)/SubjectsChunk)}
	i := 0
	for chunk := range slices.Chunk(subjects, SubjectsChunk) {
		for _, item := range chunk {
			markup.InlineKeyboard[i] = append(markup.InlineKeyboard[i], tgbotapi.NewInlineKeyboardButtonData(item, createLessonCallbackData(item)))
		}
		i++
	}
	response := tgbotapi.NewMessage(message.Chat.ID, responseText)
	response.ReplyMarkup = markup
	sentMsg, err := state.bot.SendCtx(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to send response to user: %w", err)
	}
	jsonedInfo, err := json.Marshal(DeleteStatesInfo{SentMsgId: sentMsg.MessageID})
	if err != nil {
		return fmt.Errorf("failed to marshal delete states info into json; %w", err)
	}

	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedInfo))
	if err != nil {
		return fmt.Errorf("failed to save jsoned info in delete start state: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.DELETE_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during request delete start state: %w", err)
	}
	return nil
}

func (state *DeleteStartState) Revert(ctx context.Context, msg tgbotapi.Message) error {
	return nil
}

func createLessonCallbackData(lessonName string) string {
	return constants.DELETE_REQUEST_LESSON_CALLBACK + lessonName
}
func parseLessonCallbackData(callbackData string) string {
	after, _ := strings.CutPrefix(callbackData, constants.DELETE_REQUEST_LESSON_CALLBACK)
	return after
}

type DeleteWaitingState struct {
	bot   *tgutils.Bot
	cache interfaces.HandlersCache
}

func NewDeleteWaitingState(bot *tgutils.Bot, cache interfaces.HandlersCache) *DeleteWaitingState {
	return &DeleteWaitingState{bot: bot, cache: cache}
}

func (state *DeleteWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Выберите пару для удаления заявки"))
	if err != nil {
		return fmt.Errorf("failed to send response in request delete waiting state: %w", err)
	}
	return nil
}

func (state *DeleteWaitingState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("faield to get info from cache in request delete waiting state: %w", err)
	}
	var info DeleteStatesInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info in request delete waiting state: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, info.SentMsgId, tgbotapi.NewInlineKeyboardMarkup(make([]tgbotapi.InlineKeyboardButton, 0))))
	if err != nil {
		return fmt.Errorf("failed to remove markup during delete waiting state reversal: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition state during request delete waiting state reversal: %w", err)
	}
	return nil
}

type StateMachine interface {
	Handle(ctx context.Context, message *tgbotapi.Message) error
}

type DeleteChooseState struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	machine  StateMachine
	requests RequestsRepository
}

func NewDeleteChooseState(cache interfaces.HandlersCache, bot *tgutils.Bot, machine StateMachine) *DeleteChooseState {
	return &DeleteChooseState{cache: cache, bot: bot, machine: machine}
}

func (state *DeleteChooseState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	num, err := strconv.ParseInt(message.Text, 10, 64)
	if err != nil {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите корректный номер заявки"))
		if err != nil {
			return fmt.Errorf("failed to send incorrect number response in delete request choose state: %w", err)
		}
	}
	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info in delete choose state: %w", err)
	}

	var info DeleteStatesInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info into states info: %w", err)
	}
	if num-1 < 0 || num-1 < int64(len(info.Requests)) {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Пожалуйста, введите валидное число в пределах от 1 до %d", len(info.Requests))))
		if err != nil {
			return fmt.Errorf("failed to send invalid len response in delete request choose state: %w", err)
		}
	}

	err = state.requests.Delete(ctx, info.Requests[num-1].Id)
	if err != nil {
		return fmt.Errorf("failed to delete request %d in delete request choose state: %w", info.Requests[num-1].Id, err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	return nil
}

func (state *DeleteChooseState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.RemoveInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove info from cache in delete choose state: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.DELETE_REQUEST_START))
	if err != nil {
		return fmt.Errorf("failed to save new state during delete request choose state reversal: %w", err)
	}
	return state.machine.Handle(ctx, message)
}
