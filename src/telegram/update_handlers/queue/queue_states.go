package queue

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QueueStartState struct {
	bot       *tgutils.Bot
	cache     interfaces.HandlersCache
	usersRepo interfaces.UsersRepository
	lessons   interfaces.LessonsRepository
}

func NewQueueStartState(bot *tgutils.Bot, cache interfaces.HandlersCache, usersRepo interfaces.UsersRepository, lessons interfaces.LessonsRepository) *QueueStartState {
	return &QueueStartState{
		bot:       bot,
		cache:     cache,
		usersRepo: usersRepo,
		lessons:   lessons,
	}
}

func (*QueueStartState) StateName() string {
	return constants.QUEUE_START_STATE
}

func (state *QueueStartState) Handle(ctx context.Context, msg *tgbotapi.Message) error {
	usr, err := state.usersRepo.GetByTgId(ctx, msg.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during queue command handling: %w", err)
	}
	if usr.GroupId == 0 {
		_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Вы пока не принадлежите ни к одной группе"))
		if err != nil {
			return fmt.Errorf("failed to send no group message during queue command handling: %w", err)
		}
		return nil
	}
	subjects, err := state.lessons.GetSubjects(ctx, usr.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get group subjects during queue command handling: %w", err)
	}
	if len(subjects) == 0 {
		newState := interfaces.NewCachedInfo(msg.Chat.ID, constants.IDLE_STATE)
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Больше не осталось лабораторных. Отдохните"))
		if err != nil {
			return fmt.Errorf("failed to send no labworks message during queue command handling: %w", err)
		}
		err = state.cache.SaveState(ctx, *newState)
		if err != nil {
			return fmt.Errorf("failed to save idle state during queue command handling: %w", err)
		}
		return nil
	}

	response := tgbotapi.NewMessage(msg.Chat.ID, "Выберите предмет")
	response.ReplyMarkup = createLabworksKeyboard(msg.From.ID, subjects)
	sentMsg, err := state.bot.SendCtx(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to send respponse during queue command handling: %w", err)
	}

	err = state.cache.SaveInfo(ctx, msg.Chat.ID, fmt.Sprint(sentMsg.MessageID))
	if err != nil {
		return fmt.Errorf("failed to save message id to cache during queue start state: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.QUEUE_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("failed to save queue waiting state during queue start state hadnling: %w", err)
	}
	return nil
}

func (state *QueueStartState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	return nil
}

func createLabworksKeyboard(userTgId int64, subjects []string) *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(subjects, 4) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(discipline, createQueueDisciplineCallback(userTgId, discipline)))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}

func createQueueDisciplineCallback(userTgId int64, discipline string) string {
	return constants.QUEUE_DISCIPLINE_CALLBACKS + "|" + fmt.Sprint(userTgId) + "|" + discipline
}

func parseQueueDisciplineCallback(callback string) (tgId int64, discipline string) {
	callback, _ = strings.CutPrefix(callback, constants.QUEUE_DISCIPLINE_CALLBACKS+"|")
	callbackParts := strings.Split(callback, "|")
	if len(callbackParts) != 2 {
		return 0, ""
	}
	tgId, err := strconv.ParseInt(callbackParts[0], 10, 64)
	if err != nil {
		return 0, ""
	}
	return tgId, callbackParts[1]
}

type QueueWaitingState struct {
	bot   *tgutils.Bot
	cache interfaces.HandlersCache
}

func NewQueueWaitingState(cache interfaces.HandlersCache, bot *tgutils.Bot) *QueueWaitingState {
	return &QueueWaitingState{cache: cache, bot: bot}
}

func (*QueueWaitingState) StateName() string {
	return constants.QUEUE_WAITING_STATE
}

func (*QueueWaitingState) Handle(ctx context.Context, msg *tgbotapi.Message) error {
	return nil
}

func (state *QueueWaitingState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	info, err := state.cache.GetInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info during queue waiting state reversal: %w", err)
	}
	markupMsgId, err := strconv.ParseInt(info, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse info (%s), as msg id int64: %w", info, err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, int(markupMsgId), tgbotapi.NewInlineKeyboardMarkup(make([]tgbotapi.InlineKeyboardButton, 0))))
	if err != nil {
		return fmt.Errorf("failed to send delete message during queue waiting state reversal: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during queue waiting state reversal: %w", err)
	}
	return nil
}
