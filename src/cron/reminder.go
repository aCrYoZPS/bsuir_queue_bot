package cron

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Task interface {
	Run()
}

type SheetsApi interface {
	ClearSpreadsheet(ctx context.Context, spreadsheetId string, before time.Time) error
	AddLabwork(ctx context.Context, req *labworks.AppendedLabwork) error
}

type LessonsRepo interface {
	GetEndedLessons(context.Context, time.Time) ([]persistance.Lesson, error)
	GetLessonByRequest(ctx context.Context, requestId int64) (*persistance.Lesson, error)
}

type LessonsRequestsRepository interface {
	Get(ctx context.Context, requestId int64) (*entities.LessonRequest, error)
	GetLessonRequests(ctx context.Context, lessonId int64) ([]entities.LessonRequest, error)
	Delete(ctx context.Context, requestId int64) error
	SetToNextLesson(ctx context.Context, requestId int64) error
}

type UsersRepo interface {
	GetByRequestId(ctx context.Context, requestId int64) (*entities.User, error)
}

var _ Task = (*ReminderTask)(nil)

type ReminderTask struct {
	sheets         SheetsApi
	lessons        LessonsRepo
	lessonsRequest LessonsRequestsRepository
	users          UsersRepo
	bot            *tgutils.Bot
}

func NewReminderTask(sheets SheetsApi, lessons LessonsRepo, lessonsRequest LessonsRequestsRepository, users UsersRepo, bot *tgutils.Bot) *ReminderTask {
	return &ReminderTask{sheets: sheets, lessons: lessons, lessonsRequest: lessonsRequest, users: users, bot: bot}
}

const SHEETS_REFRESH_TIMEOUT = 30 * time.Second

func (task *ReminderTask) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), SHEETS_REFRESH_TIMEOUT)
	defer cancel()
	endedLessons, err := task.lessons.GetEndedLessons(ctx, time.Now().AddDate(0, 0, 0))
	if err != nil {
		slog.Error(fmt.Errorf("failed getting ended lessons for reminder task: %w", err).Error())
	}

	requests := make([]entities.LessonRequest, 0, len(endedLessons))
	for _, lesson := range endedLessons {
		storedRequests, err := task.lessonsRequest.GetLessonRequests(ctx, lesson.GroupId)
		if err != nil {
			slog.Error(fmt.Errorf("failed getting lessons request for reminder task: %w", err).Error())
		}
		requests = append(requests, storedRequests...)
	}

	for _, request := range requests {
		err := task.sendMessageForRequested(ctx, &request)
		if err != nil {
			slog.Error(fmt.Errorf("failed to send msg to user id %d during reminder task: %w", request.UserId, err).Error())
		}
	}

}

func (task *ReminderTask) sendMessageForRequested(ctx context.Context, request *entities.LessonRequest) error {
	msg := tgbotapi.NewMessage(request.UserId, "Вы сдали данную лабораторную?")
	msg.ReplyToMessageID = int(request.MsgId)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Да", createSheetsRefreshCallbackData(request.Id, true)),
		tgbotapi.NewInlineKeyboardButtonData("Нет", createSheetsRefreshCallbackData(request.Id, true))})
	_, err := task.bot.SendCtx(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}

const REMINDER_CALLBACKS = "sheets_refr"

func createSheetsRefreshCallbackData(requestId int64, accepted bool) string {
	formattedAccepted := 0
	if accepted {
		formattedAccepted = 1
	}
	return REMINDER_CALLBACKS + "|" + fmt.Sprint(formattedAccepted) + "|" + fmt.Sprint(requestId)
}

func parseSheetsRefreshCallbackData(callbackData string) (requestId int64, accepted bool) {
	callbackData, _ = strings.CutPrefix(callbackData, REMINDER_CALLBACKS+"|")
	formattedAccepted, formattedRequestId, _ := strings.Cut(callbackData, "|")
	accepted = false
	acceptedInt, _ := strconv.Atoi(formattedAccepted)
	if acceptedInt == 1 {
		accepted = true
	}
	requestId, _ = strconv.ParseInt(formattedRequestId, 10, 64)
	return requestId, accepted
}
