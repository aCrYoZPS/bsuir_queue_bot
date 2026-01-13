package queue

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	customErrors "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/errors"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LabworksRequest interface {
	GetLabworkQueue(ctx context.Context, labworkId int64) ([]entities.User, error)
}

type QueueCallbacksHandler struct {
	users    interfaces.UsersRepository
	labworks interfaces.LessonsRepository
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	requests LabworksRequest
}

func NewQueueCallbackHandler(users interfaces.UsersRepository, labworks interfaces.LessonsRepository, cache interfaces.HandlersCache, bot *tgutils.Bot, requests LabworksRequest) *QueueCallbacksHandler {
	return &QueueCallbacksHandler{
		users:    users,
		labworks: labworks,
		cache:    cache,
		bot:      bot,
		requests: requests,
	}
}

func (handler *QueueCallbacksHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	if strings.HasPrefix(update.CallbackData(), constants.QUEUE_DISCIPLINE_CALLBACKS) {
		userTgId, discipline := parseQueueDisciplineCallback(update.CallbackData())
		if discipline == "" || userTgId == 0 {
			return errors.New("invalid command requested")
		}
		err := handler.handleDisciplineCallback(ctx, update.CallbackQuery.Message, discipline)
		if err != nil {
			if errors.Is(err, customErrors.ErrNoLabworks) {
				err := handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.CallbackQuery.Message.Chat.ID, constants.IDLE_STATE))
				if err != nil {
					return fmt.Errorf("failed to transition to idle state during labwork callback handling: %w", err)
				}
				_, err = bot.SendCtx(ctx, tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, cases.Title(language.English).String(customErrors.ErrNoLabworks.Error())))
				return fmt.Errorf("failed to send no lessons error to user during labwork callback handling: %w", err)
			}
			return err
		}
		return nil
	} else if strings.HasPrefix(update.CallbackData(), constants.QUEUE_TIME_CALLBACKS) {
		err := handler.handleTimeCallback(ctx, update)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(update.CallbackData(), constants.QUEUE_CANCEL_CALLBACKS) {
		err := handler.handleCancelCallback(ctx, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (handler *QueueCallbacksHandler) handleCancelCallback(ctx context.Context, update *tgbotapi.Update) error {
	info, err := handler.cache.GetInfo(ctx, update.FromChat().ID)
	if err != nil {
		return fmt.Errorf("failed to get info during queue cancel callback handling: %w", err)
	}
	msgId, err := strconv.Atoi(info)
	if err != nil {
		return fmt.Errorf("failed to get msg id during queue cancel callback handling: %w", err)
	}
	usr, err := handler.users.GetByTgId(ctx, update.SentFrom().ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during queue cancel callback handling: %w", err)
	}
	subjects, err := handler.labworks.GetSubjects(ctx, usr.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get subjects during queue cancel callback handling: %w", err)
	}
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.FromChat().ID, msgId, *createLabworksKeyboard(usr.TgId, subjects)))
	if err != nil {
		return fmt.Errorf("failed to send new reply markup durning queue cancel callback handling: %w", err)
	}
	return nil
}

func (handler *QueueCallbacksHandler) handleDisciplineCallback(ctx context.Context, msg *tgbotapi.Message, discipline string) error {
	user, err := handler.users.GetByTgId(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during labworks discipline callback: %w", err)
	}
	lessons, err := handler.labworks.GetNext(ctx, discipline, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get next labworks during labworks discipline callback: %w", err)
	}
	if lessons == nil {
		return customErrors.ErrNoLabworks
	}

	keyboard := handler.createDisciplineDatesKeyboard(lessons)
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, *keyboard))
	if err != nil {
		return fmt.Errorf("failed to send keyboard during labworks callback handling: %w", err)
	}
	return nil
}

func (handler *QueueCallbacksHandler) createDisciplineDatesKeyboard(lessons []persistance.Lesson) *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for _, lesson := range lessons {
		row := []tgbotapi.InlineKeyboardButton{}
		formattedDate := fmt.Sprintf("%02d/%02d/%d", lesson.DateTime.Day(), lesson.DateTime.Month(), lesson.DateTime.Year())
		if lesson.SubgroupNumber != iis_api_entities.AllSubgroups {
			formattedDate += fmt.Sprintf(" (%d)", lesson.SubgroupNumber)
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(formattedDate, createQueueTimeCallback(lesson.Id, lesson.DateTime, lesson.SubgroupNumber)))
		markup = append(markup, row)
	}
	markup = append(markup, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", constants.QUEUE_CANCEL_CALLBACKS)))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}

func createQueueTimeCallback(lessonId int64, date time.Time, subgroup iis_api_entities.Subgroup) string {
	builder := strings.Builder{}
	builder.Grow(64)
	builder.WriteString(constants.QUEUE_TIME_CALLBACKS)
	builder.WriteString("|")
	builder.WriteString(fmt.Sprintf("%d.%d.%d", date.Day(), date.Month(), date.Year()))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(lessonId))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(subgroup))
	return builder.String()
}

func parseLabworkTimeCallback(callback string) (date time.Time, labworkId int64, subgroup int8) {
	callback, _ = strings.CutPrefix(callback, constants.QUEUE_TIME_CALLBACKS+"|")
	subgroup = iis_api_entities.AllSubgroups
	formattedDate, after, _ := strings.Cut(callback, "|")

	after, subgroupString, _ := strings.Cut(after, "|")
	labworkId, err := strconv.ParseInt(after, 10, 64)
	if err != nil {
		return time.Time{}, 0, subgroup
	}
	nums := strings.Split(formattedDate, ".")
	if len(nums) < 3 {
		return time.Time{}, 0, subgroup
	}
	day, err := strconv.Atoi(nums[0])
	if err != nil {
		return time.Time{}, 0, subgroup
	}
	month, err := strconv.Atoi(nums[1])
	if err != nil {
		return time.Time{}, 0, subgroup
	}
	year, err := strconv.Atoi(nums[2])
	if err != nil {
		return time.Time{}, 0, subgroup
	}
	date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if subgroupString != "" {
		subgroupVal, err := strconv.Atoi(subgroupString)
		if err != nil {
			return time.Time{}, 0, subgroup
		}
		subgroup = int8(subgroupVal)
	}
	return date, labworkId, subgroup
}

func (handler *QueueCallbacksHandler) handleTimeCallback(ctx context.Context, update *tgbotapi.Update) error {
	_, labworkId, _ := parseLabworkTimeCallback(update.CallbackData())
	users, err := handler.requests.GetLabworkQueue(ctx, labworkId)
	if err != nil {
		return fmt.Errorf("failed to get labwork queue from db: %w", err)
	}
	var output strings.Builder
	for i, user := range users {
		output.WriteString(fmt.Sprintf("%d %s\n", i+1, user.FullName))
	}
	if output.String() == "" {
		output.WriteString("На эту лабораторную нет заявок. Как знать,может,вы будете первым")
	}
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageTextAndMarkup(update.FromChat().ChatConfig().ChatID, update.CallbackQuery.Message.MessageID, output.String(),
		tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)}))
	if err != nil {
		return fmt.Errorf("failed to send queue during queue time callback handling: %w", err)
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.FromChat().ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during queue time callback handling: %w", err)
	}
	return nil
}
