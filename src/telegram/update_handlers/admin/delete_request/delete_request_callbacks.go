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
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LessonsRepository interface {
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
	GetNext(ctx context.Context, subject string, groupId int64) ([]persistance.Lesson, error)
}

type DeleteLessonCallbackHandler struct {
	cache   interfaces.HandlersCache
	users   UsersRepository
	bot     *tgutils.Bot
	lessons LessonsRepository
}

func NewDeleteLessonCallbackHandler(cache interfaces.HandlersCache, users UsersRepository, bot *tgutils.Bot, lessons LessonsRepository) *DeleteLessonCallbackHandler {
	return &DeleteLessonCallbackHandler{
		cache:   cache,
		users:   users,
		bot:     bot,
		lessons: lessons,
	}
}

func (state *DeleteLessonCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	lessonName := parseLessonCallbackData(update.CallbackData())
	user, err := state.users.GetByTgId(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user in delete lesson reequest callback: %w", err)
	}
	lessons, err := state.lessons.GetNext(ctx, lessonName, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get next lessons in delete lesson request callback: %w", err)
	}

	jsonedInfo, err := state.cache.GetInfo(ctx, update.FromChat().ChatConfig().ChatID)
	if err != nil {
		return fmt.Errorf("failed to get info in delete lesson request callback: %w", err)
	}

	var info DeleteStatesInfo

	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json into info in delete lesson request callback: %w", err)
	}

	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.FromChat().ChatConfig().ChatID, info.SentMsgId, state.createMarkup(lessons)))
	if err != nil {
		return fmt.Errorf("failed to edit markup in delete lesson request callback: %w", err)
	}
	return nil
}

func (state *DeleteLessonCallbackHandler) createMarkup(lessons []persistance.Lesson) tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	row := []tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(lessons, 4) {
		for _, lesson := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(lesson.Subject+fmt.Sprintf(" (%s)",
				lesson.DateTime.Format("02.01.2006")), createTimeCallbackData(lesson.Id)))
		}
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	return keyboard
}

func createTimeCallbackData(lessonId int64) string {
	return constants.DELETE_REQUEST_TIME_CALLBACK + "|" + fmt.Sprint(lessonId)
}

func parseTimeCallbackData(callbackData string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(callbackData, constants.DELETE_REQUEST_TIME_CALLBACK+"|"), 10, 64)
}

type LessonRequestsRepository interface {
	GetLessonRequests(ctx context.Context, lessonId int64) ([]entities.LessonRequest, error)
}

type DeleteTimeCallbackHandler struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	requests LessonRequestsRepository
	users    UsersRepository
}

func NewDeleteTimeCallbackHandler(cache interfaces.HandlersCache, bot *tgutils.Bot, requests LessonRequestsRepository, users UsersRepository) *DeleteTimeCallbackHandler {
	return &DeleteTimeCallbackHandler{cache: cache, bot: bot, requests: requests, users: users}
}

func (state *DeleteTimeCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	lessonId, err := parseTimeCallbackData(update.CallbackData())
	if err != nil {
		return fmt.Errorf("failed to parse callback data: %w", err)
	}

	requests, err := state.requests.GetLessonRequests(ctx, lessonId)
	if err != nil {
		return fmt.Errorf("failed to get requests in delete request time callback handler: %w", err)
	}
	users := make([]entities.User, len(requests))
	for _, request := range requests {
		user, err := state.users.Get(ctx, request.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user by id in delete request time callback handler: %w", err)
		}
		users = append(users, user)
	}

	jsonedInfo, err := json.Marshal(&DeleteStatesInfo{Requests: requests})
	if err != nil {
		return fmt.Errorf("failed to convert info into json in request delete time callback handler: %w", err)
	}
	err = state.cache.SaveInfo(ctx, update.FromChat().ID, string(jsonedInfo))
	if err != nil {
		return fmt.Errorf("failed to save info in delete request time callback handler: %w", err)
	}

	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(update.FromChat().ID, formatOutput(requests, users)))
	if err != nil {
		return fmt.Errorf("failed to send response to user in delete request time callback handler: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.FromChat().ID, constants.DELETE_REQUEST_CHOOSE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save delete choose state in delete request time callback handler: %w", err)
	}
	return nil
}

func formatOutput(requests []entities.LessonRequest, users []entities.User) string {
	var out strings.Builder
	fmt.Fprintf(&out, "Введите число, представляющее заявку, для удаления\n")
	for i := range requests {
		fmt.Fprintf(&out, "%d. %s, лабораторная: %d\n", i+1, users[i].FullName, requests[i].LabworkNumber)
	}
	return out.String()
}
