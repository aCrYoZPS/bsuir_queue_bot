package cron

import (
	"context"
	"fmt"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/labworks"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReminderCallbackHandler struct {
	lessons         LessonsRepoReminder
	lessonsRequests LessonsRequestsRepositoryReminder
	sheets          SheetsApiReminder
	users           UsersRepoReminder
}

func NewSheetsRefreshCallbackHandler(lessonsRequests LessonsRequestsRepositoryReminder, sheets SheetsApiReminder, users UsersRepoReminder, lessons LessonsRepoReminder) *ReminderCallbackHandler {
	return &ReminderCallbackHandler{lessonsRequests: lessonsRequests, sheets: sheets, users: users, lessons: lessons}
}

func (handler *ReminderCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	if strings.HasPrefix(update.CallbackData(), REMINDER_CALLBACKS) {
		requestId, accepted := parseSheetsRefreshCallbackData(update.CallbackData())
		if accepted {
			err := handler.lessonsRequests.Delete(ctx, requestId)
			if err != nil {
				return fmt.Errorf("failed to delete lesson request in sheets refresh: %w", err)
			}
		} else {
			err := handler.SetNextLesson(ctx, requestId)
			if err != nil {
				return err
			}
		}
		_, err := bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.FromChat().ID, update.CallbackQuery.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return fmt.Errorf("failed to delete reply markup on a reminder message: %w", err)
		}
	} else {
		return fmt.Errorf("wrong callback data (%s) passed to sheets refresh callback handler", update.CallbackData())
	}
	return nil
}

func (handler *ReminderCallbackHandler) SetNextLesson(ctx context.Context, requestId int64) error {
	err := handler.lessonsRequests.SetToNextLesson(ctx, requestId)
	if err != nil {
		return fmt.Errorf("failed to set lesson request to next lesson in sheets refresh cron: %w", err)
	}
	usr, err := handler.users.GetByRequestId(ctx, requestId)
	if err != nil {
		return fmt.Errorf("failed to get user by request id in sheets refresh cron: %w", err)
	}
	lesson, err := handler.lessons.GetLessonByRequest(ctx, requestId)
	if err != nil {
		return fmt.Errorf("failed to get lessons by request id in sheets refresh cron: %w", err)
	}
	req, err := handler.lessonsRequests.Get(ctx, requestId)
	if err != nil {
		return fmt.Errorf("failed to get lesson request by id in sheets refresh cron: %w", err)
	}
	err = handler.sheets.AddLabworkRequest(ctx, labworks.NewAppendedLabwork(lesson.DateTime, req.SubmitTime, lesson.Subject, usr.GroupName, usr.FullName, lesson.SubgroupNumber, req.LabworkNumber))
	if err != nil {
		return fmt.Errorf("failed to add labwork to sheets during sheets refresh cron: %w", err)
	}
	return nil
}
