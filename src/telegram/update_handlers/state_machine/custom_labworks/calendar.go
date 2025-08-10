package customlabworks

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	additionalRows = 3
	calendarStart  = 2
)

var months = map[time.Month]string{
	time.January:   "Январь",
	time.February:  "Февраль",
	time.March:     "Марть",
	time.April:     "Апрель",
	time.May:       "Май",
	time.June:      "Июнь",
	time.July:      "Июль",
	time.August:    "Август",
	time.September: "Сентябрь",
	time.October:   "Октябрь",
	time.November:  "Ноябрь",
	time.December:  "Декабрь",
}

var days = map[time.Weekday]string{
	time.Monday:    "Пн",
	time.Tuesday:   "Вт",
	time.Wednesday: "Ср",
	time.Thursday:  "Чт",
	time.Friday:    "Пт",
	time.Saturday:  "Сб",
	time.Sunday:    "Вс",
}

const (
	daysInWeek = 7
)

const (
	oneSideNavigationSize = 2
	twoSideNavigationSize = 3
)

func createCalendar(date time.Time, isCurrentMonth bool) *tgbotapi.InlineKeyboardMarkup {
	currentYear, currentMonth, currentDay := date.Date()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.Local)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	_, lastOfMonthWeek := lastOfMonth.ISOWeek()
	_, firstOfMonthWeek := firstOfMonth.ISOWeek()

	markup := make([][]tgbotapi.InlineKeyboardButton, int(math.Abs(float64(lastOfMonthWeek-firstOfMonthWeek+1)))+additionalRows)

	createCalendarHeader(&markup, currentMonth, currentYear)
	createDateRows(&markup, firstOfMonth.Weekday(), currentDay, int(currentMonth), currentYear, lastOfMonth)

	if isCurrentMonth {
		markup[len(markup)-1] = make([]tgbotapi.InlineKeyboardButton, oneSideNavigationSize)
		markup[len(markup)-1][0] = tgbotapi.NewInlineKeyboardButtonData(" ", constants.IGNORE_CALLBACKS)
		markup[len(markup)-1][1] = tgbotapi.NewInlineKeyboardButtonData(">>", createForwardCallback(currentDay, int(currentMonth), currentYear))
	} else {
		markup[len(markup)-1] = make([]tgbotapi.InlineKeyboardButton, twoSideNavigationSize)
		markup[len(markup)-1][0] = tgbotapi.NewInlineKeyboardButtonData("<<", createBackCallback(currentDay, int(currentMonth), currentYear))
		markup[len(markup)-1][1] = tgbotapi.NewInlineKeyboardButtonData(" ", constants.IGNORE_CALLBACKS)
		markup[len(markup)-1][2] = tgbotapi.NewInlineKeyboardButtonData(">>", createForwardCallback(currentDay, int(currentMonth), currentYear))
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: markup}
}

func createDateRows(markup *[][]tgbotapi.InlineKeyboardButton, firstDayWeekday time.Weekday, currentDay, currentMonth, currentYear int, lastOfMonth time.Time) {
	for i := range 7 {
		(*markup)[1][i] = tgbotapi.NewInlineKeyboardButtonData(days[time.Weekday(i)], constants.IGNORE_CALLBACKS)
	}

	displayedDate := 1
	for i := calendarStart; i < len(*markup)-1; i++ {
		j := 0
		if i == calendarStart {
			for j = 0; j < int(firstDayWeekday); j++ {
				(*markup)[i][j] = tgbotapi.NewInlineKeyboardButtonData("-", constants.IGNORE_CALLBACKS)
			}
			j = int(firstDayWeekday)
		}
		for ; j < 7; j, displayedDate = j+1, displayedDate+1 {
			if displayedDate >= currentDay && displayedDate <= lastOfMonth.Day() {
				(*markup)[i][j] = tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(displayedDate), createDateCallback(displayedDate, currentMonth, currentYear))
			} else {
				(*markup)[i][j] = tgbotapi.NewInlineKeyboardButtonData("-", constants.IGNORE_CALLBACKS)
			}
		}
	}
}

func createCalendarHeader(markup *[][]tgbotapi.InlineKeyboardButton, currentMonth time.Month, currentYear int) {
	currentMonthString := fmt.Sprintf(months[currentMonth]+" %d", currentYear)
	(*markup)[0] = make([]tgbotapi.InlineKeyboardButton, 1)
	(*markup)[0] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(currentMonthString, constants.IGNORE_CALLBACKS))

	for i := calendarStart; i < len(*markup)-1; i++ {
		(*markup)[i] = make([]tgbotapi.InlineKeyboardButton, daysInWeek)
	}

	(*markup)[1] = make([]tgbotapi.InlineKeyboardButton, 7)
	//Row of week days
	for i := range 7 {
		(*markup)[1][i] = tgbotapi.NewInlineKeyboardButtonData(days[time.Weekday(i)], constants.IGNORE_CALLBACKS)
	}
}

func createDateCallback(displayedDate, currentMonth, currentYear int) string {
	return constants.CALENDAR_DATE_CALLBACKS + createEuropeanDate(displayedDate, currentMonth, currentYear)
}

func parseDateCallback(callback string) time.Time {
	europeanDate, _ := strings.CutPrefix(callback, constants.CALENDAR_DATE_CALLBACKS)
	return parseEuropeanDate(europeanDate)
}

func createForwardCallback(displayedDate, currentMonth, currentYear int) string {
	return constants.CALENDAR_NAVIGATE_FRONT_CALLBACK + createEuropeanDate(displayedDate, currentMonth, currentYear)
}

func parseForwardCallback(callback string) time.Time {
	europeanDate, _ := strings.CutPrefix(callback, constants.CALENDAR_NAVIGATE_FRONT_CALLBACK)
	return parseEuropeanDate(europeanDate)
}

func createBackCallback(displayedDate, currentMonth, currentYear int) string {
	return constants.CALENDAR_NAVIGATE_BACK_CALLBACK + createEuropeanDate(displayedDate, currentMonth, currentYear)
}

func parseBackCallback(callback string) time.Time {
	europeanDate, _ := strings.CutPrefix(callback, constants.CALENDAR_NAVIGATE_BACK_CALLBACK)
	return parseEuropeanDate(europeanDate)
}

func createEuropeanDate(displayedDate, currentMonth, currentYear int) string {
	return fmt.Sprintf("%d.%d.%d", displayedDate, currentMonth, currentYear)
}

func parseEuropeanDate(europeanDate string) time.Time {
	date := strings.Split(europeanDate, ".")
	if len(date) != 3 {
		return time.Time{}
	}
	year, err := strconv.Atoi(date[2])
	if err != nil {
		return time.Time{}
	}
	month, err := strconv.Atoi(date[1])
	if err != nil {
		return time.Time{}
	}
	day, err := strconv.Atoi(date[0])
	if err != nil {
		return time.Time{}
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

type CalendarCallbackHandler struct {
	bot   *tgutils.Bot
	cache interfaces.HandlersCache
}

func NewCalendarCallbackHandler(bot *tgutils.Bot, cache interfaces.HandlersCache) *CalendarCallbackHandler {
	return &CalendarCallbackHandler{
		bot:   bot,
		cache: cache,
	}
}

func (handler *CalendarCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	var err error
	switch {
	case strings.HasPrefix(update.CallbackData(), constants.CALENDAR_NAVIGATE_FRONT_CALLBACK):
		err = handler.handleNavigateFront(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.CALENDAR_NAVIGATE_BACK_CALLBACK):
		err = handler.handleNavigateBack(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.CALENDAR_DATE_CALLBACKS):
		err = handler.handleDate(ctx, update)
	case strings.HasPrefix(update.CallbackData(), constants.IGNORE_CALLBACKS):

	default:
		err = fmt.Errorf("failed to get matching callback for string (%s) when handling calendar callbacks", update.CallbackData())
	}
	return err
}

func (handler *CalendarCallbackHandler) handleNavigateFront(ctx context.Context, update *tgbotapi.Update) error {
	curDate := parseForwardCallback(update.CallbackData())
	curDate = curDate.AddDate(0, 1, -curDate.Day()+1)

	_, err := handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createCalendar(curDate, false)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup while navigating front in calendar: %w", err)
	}
	return nil
}

func (handler *CalendarCallbackHandler) handleNavigateBack(ctx context.Context, update *tgbotapi.Update) error {
	isCurrentMonth := false
	curDate := parseBackCallback(update.CallbackData())
	curDate = curDate.AddDate(0, -1, -curDate.Day()+1)
	if curDate.Month() == time.Now().Month() {
		curDate = time.Now()
		isCurrentMonth = true
	}
	_, err := handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createCalendar(curDate, isCurrentMonth)))
	if err != nil {
		return fmt.Errorf("failed to edit reply markup while navigating back in calendar: %w", err)
	}
	return nil
}

func (handler *CalendarCallbackHandler) handleDate(ctx context.Context, update *tgbotapi.Update) error {
	curDate := parseDateCallback(update.CallbackData())
	jsonedInfo, err := handler.cache.GetInfo(ctx, update.CallbackQuery.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info during date callback handling in calendar: %w", err)
	}
	req := &LessonRequest{}
	if err = json.Unmarshal([]byte(jsonedInfo), req); err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info into lesson request during date callback handling in calendar: %w", err)
	}
	req.DateTime = curDate

	//Can't return an error if you could unmarshal it
	jsonBytes, _ := json.Marshal(req)
	err = handler.cache.SaveInfo(ctx, update.CallbackQuery.Message.Chat.ID, string(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to save jsoned info during date callback handling in calendar: %w", err)
	}

	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, *createTimePicker("")))
	if err != nil {
		return fmt.Errorf("failed to edit message during date callback handling in calendar: %w", err)
	}
	return nil
}
