package customlabworks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sheetsapi "github.com/aCrYoZPS/bsuir_queue_bot/src/google_docs/sheets_api"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/googleapi"
)

// Creates time picker markup with the given string in format "15:00" as time. Pass empty string to set default time
func createTimePicker(currentTime string) *tgbotapi.InlineKeyboardMarkup {
	if currentTime == "" {
		currentTime = "15:00"
	}
	markup := make([][]tgbotapi.InlineKeyboardButton, 5)
	markup[0] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(currentTime, constants.IGNORE_CALLBACKS)}
	markup[1] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("+", createHoursIncreaseCallback(currentTime)),
		tgbotapi.NewInlineKeyboardButtonData("-", createHoursDecreaseCallback(currentTime))}
	markup[2] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("+", createMinutesIncreaseCallback(currentTime)),
		tgbotapi.NewInlineKeyboardButtonData("-", createMinutesDecreaseCallback(currentTime))}
	markup[3] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Назад", constants.TIME_CANCEL)}
	markup[4] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Принять", createTimeAcceptCallback(currentTime))}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: markup}
}

func createHoursIncreaseCallback(currentTime string) string {
	return constants.TIME_HOURS_INCREASE_CALLBACKS + currentTime
}

func parseHoursIncreaseCallback(callbackData string) string {
	after, _ := strings.CutPrefix(callbackData, constants.TIME_HOURS_INCREASE_CALLBACKS)
	return after
}

func createMinutesIncreaseCallback(currentTime string) string {
	return constants.TIME_MINUTES_INCREASE_CALLBACKS + currentTime
}

func parseMinutesIncreaseCallback(callbackData string) string {
	after, _ := strings.CutPrefix(callbackData, constants.TIME_MINUTES_INCREASE_CALLBACKS)
	return after
}

func createHoursDecreaseCallback(currentTime string) string {
	return constants.TIME_HOURS_DECREASE_CALLBACKS + currentTime
}

func parseHoursDecreaseCallback(callbackData string) string {
	after, _ := strings.CutPrefix(callbackData, constants.TIME_HOURS_DECREASE_CALLBACKS)
	return after
}

func createMinutesDecreaseCallback(currentTime string) string {
	return constants.TIME_MINUTES_DESCREASE_CALLBACKS + currentTime
}

func parseMinutesDecreaseCallback(callbackData string) string {
	after, _ := strings.CutPrefix(callbackData, constants.TIME_MINUTES_DESCREASE_CALLBACKS)
	return after
}

func createTimeAcceptCallback(currentTime string) string {
	return constants.TIME_SUBMIT + currentTime
}

func parseTimeAcceptCallback(callback string) time.Time {
	callback, _ = strings.CutPrefix(callback, constants.TIME_SUBMIT)
	acceptedTime, _ := time.Parse("15:04", callback)
	return acceptedTime
}

type LessonsRepository interface {
	Add(ctx context.Context, lesson *persistance.Lesson) error
}

type TimePickerCallbackHandler struct {
	bot     *tgutils.Bot
	lessons LessonsRepository
	cache   interfaces.HandlersCache
}

func NewTimePickerCallbackHandler(bot *tgutils.Bot, lessons LessonsRepository, cache interfaces.HandlersCache) *TimePickerCallbackHandler {
	return &TimePickerCallbackHandler{bot: bot, lessons: lessons, cache: cache}
}

func (callbackHandler *TimePickerCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	var err error
	switch {
	case strings.HasPrefix(update.CallbackData(), constants.TIME_CANCEL):
		err = callbackHandler.handleCancelCallback(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_HOURS_INCREASE_CALLBACKS):
		err = callbackHandler.handleHoursIncreaseCallback(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_HOURS_DECREASE_CALLBACKS):
		err = callbackHandler.handleHoursDecreaseCallback(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_MINUTES_INCREASE_CALLBACKS):
		err = callbackHandler.handleMinutesIncreaseCallback(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_MINUTES_DESCREASE_CALLBACKS):
		err = callbackHandler.handleMinutesDecreaseCallback(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.TIME_SUBMIT):
		err = callbackHandler.handleTimeSubmitCallback(ctx, update)
	default:
		err = fmt.Errorf("failed to match callback data (%s) to callback handler in time picker callback handler", update.CallbackData())
	}
	return err
}

func (callbackHandler *TimePickerCallbackHandler) handleCancelCallback(ctx context.Context, update *tgbotapi.Update) error {
	_, err := callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createCalendar(time.Now(), true)))
	if err != nil {
		return fmt.Errorf("failed to edit markup when handling time picker cancel callback: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) handleHoursIncreaseCallback(ctx context.Context, update *tgbotapi.Update) error {
	curTimeString := parseHoursIncreaseCallback(update.CallbackData())

	hoursAndMinutes := strings.Split(curTimeString, ":")
	if len(hoursAndMinutes) != 2 {
		return fmt.Errorf("failed to parse time string (%s) when handling time picker hours increase callback", curTimeString)
	}
	hours, err := strconv.Atoi(hoursAndMinutes[0])
	if err != nil {
		return fmt.Errorf("failed to convert hours to integer when handling time picker hours increase callback: %w", err)
	}
	minutes, err := strconv.Atoi(hoursAndMinutes[1])
	if err != nil {
		return fmt.Errorf("failed to convert minutes to integer when handling time picker hours increase callback: %w", err)
	}

	hours = (hours + 1) % 24
	curTimeString = fmt.Sprintf("%d:%02d", hours, minutes)

	_, err = callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createTimePicker(curTimeString)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup during handling time picker hours increase callback: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) handleHoursDecreaseCallback(ctx context.Context, update *tgbotapi.Update) error {
	curTimeString := parseHoursDecreaseCallback(update.CallbackData())

	hoursAndMinutes := strings.Split(curTimeString, ":")
	if len(hoursAndMinutes) != 2 {
		return fmt.Errorf("failed to parse time string (%s) when handling time picker hours decrease callback", curTimeString)
	}
	hours, err := strconv.Atoi(hoursAndMinutes[0])
	if err != nil {
		return fmt.Errorf("failed to convert hours to integer when handling time picker hours decrease callback: %w", err)
	}
	minutes, err := strconv.Atoi(hoursAndMinutes[1])
	if err != nil {
		return fmt.Errorf("failed to convert minutes to integer when handling time picker hours decrease callback: %w", err)
	}

	hours = (24 + hours - 1) % 24
	curTimeString = fmt.Sprintf("%d:%02d", hours, minutes)

	_, err = callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createTimePicker(curTimeString)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup during handling time picker hours decrease callback: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) handleMinutesIncreaseCallback(ctx context.Context, update *tgbotapi.Update) error {
	curTimeString := parseMinutesIncreaseCallback(update.CallbackData())

	hoursAndMinutes := strings.Split(curTimeString, ":")
	if len(hoursAndMinutes) != 2 {
		return fmt.Errorf("failed to parse time string (%s) when handling time picker minutes increase callback", curTimeString)
	}
	hours, err := strconv.Atoi(hoursAndMinutes[0])
	if err != nil {
		return fmt.Errorf("failed to convert hours to integer when handling time picker minutes increase callback: %w", err)
	}
	minutes, err := strconv.Atoi(hoursAndMinutes[1])
	if err != nil {
		return fmt.Errorf("failed to convert minutes to integer when handling time picker minutes increase callback: %w", err)
	}

	minutes = (minutes + 5) % 60
	curTimeString = fmt.Sprintf("%d:%02d", hours, minutes)

	_, err = callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createTimePicker(curTimeString)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup during handling time picker minutes increase callback: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) handleMinutesDecreaseCallback(ctx context.Context, update *tgbotapi.Update) error {
	curTimeString := parseMinutesDecreaseCallback(update.CallbackData())

	hoursAndMinutes := strings.Split(curTimeString, ":")
	if len(hoursAndMinutes) != 2 {
		return fmt.Errorf("failed to parse time string (%s) when handling time picker minutes decrease callback", curTimeString)
	}
	hours, err := strconv.Atoi(hoursAndMinutes[0])
	if err != nil {
		return fmt.Errorf("failed to convert hours to integer when handling time picker minutes decrease callback: %w", err)
	}
	minutes, err := strconv.Atoi(hoursAndMinutes[1])
	if err != nil {
		return fmt.Errorf("failed to convert minutes to integer when handling time picker minutes decrease callback: %w", err)
	}

	minutes = (60 + minutes - 5) % 60
	curTimeString = fmt.Sprintf("%d:%02d", hours, minutes)

	_, err = callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createTimePicker(curTimeString)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup during handling time picker minutes decrease callback: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) handleTimeSubmitCallback(ctx context.Context, update *tgbotapi.Update) error {
	selectedTime := parseTimeAcceptCallback(update.CallbackData())
	jsonedInfo, err := callbackHandler.cache.GetInfo(ctx, update.CallbackQuery.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info during handling time submit callback: %w", err)
	}
	request := &LessonRequest{}
	err = json.Unmarshal([]byte(jsonedInfo), &request)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info into lesson request during handling time submit callback: %w", err)
	}
	request.DateTime = request.DateTime.Add(selectedTime.Sub(selectedTime.Truncate(24 * time.Hour)))

	err = callbackHandler.lessons.Add(ctx, persistance.NewPersistedLesson(request.GroupId, iis_api_entities.AllSubgroups, iis_api_entities.Labwork, request.Name, request.DateTime))
	if err != nil {
		return callbackHandler.wrapLessonsServiceError(ctx, err, update)
	}
	_, err = callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Ваша лабораторная была сохранена", tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove markup from message during time picker submit callback hadnling: %w", err)
	}

	err = callbackHandler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.CallbackQuery.Message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to idle state in time submit callback handling: %w", err)
	}
	return nil
}

func (callbackHandler *TimePickerCallbackHandler) wrapLessonsServiceError(ctx context.Context, err error, update *tgbotapi.Update) error {
	if errors.Is(err, sheetsapi.ErrSheetsExists()) {
		_, err := callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID,
			"Пара под данным именем и датой уже существует. Пожалуйста, укажите другое имя", tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return fmt.Errorf("failed to send sheet exists message during time picker submit callback handling: %w", err)
		}
		err = callbackHandler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.CallbackQuery.Message.Chat.ID, constants.LABWORK_ADD_SUBMIT_NAME_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition to custom labwork name submit during time picker submit: %w", err)
		}
		return nil
	} else if errors.Is(err, sheetsapi.ErrNoSheetCreated()) {
		_, err := callbackHandler.bot.SendCtx(ctx, tgbotapi.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID,
			"Пара под данным именем и датой не создалась. Пожалуйста,проверьте,что имя валидно для названия в гугл таблицах", tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return fmt.Errorf("failed to send sheet exists message during time picker submit callback handling: %w", err)
		}
		err = callbackHandler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.CallbackQuery.Message.Chat.ID, constants.LABWORK_ADD_SUBMIT_NAME_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition to custom labwork name submit during time picker submit: %w", err)
		}
		return nil
	} else if googleErr, ok := err.(*googleapi.Error); ok {
		if googleErr.Code == http.StatusInternalServerError {
			_, err := callbackHandler.bot.SendCtx(ctx, tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка на стороне сервисов гугла. Пожалуйста,попробуйте позже"))
			if err != nil {
				return fmt.Errorf("failed to send google service error response to user during time picker callback handling: %w", err)
			}
		}
		return fmt.Errorf("failed to add custom lesson during time picker submit callback handling: %w", err)
	}
	return err
}
