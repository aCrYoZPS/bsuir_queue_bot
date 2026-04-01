package reorder

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReorderLessonCallbackHandler struct {
	bot   *tgutils.Bot
	cache interfaces.HandlersCache
}

func NewReorderLessonCallbackHandler(bot *tgutils.Bot) *ReorderLessonCallbackHandler {
	return &ReorderLessonCallbackHandler{
		bot: bot,
	}
}

func (handler *ReorderLessonCallbackHandler) Handle(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	subject := parseLessonCallback(update.CallbackData())
	if subject == "" {
		_, err := bot.SendCtx(ctx, tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Выберите корректное название предмета"))
		if err != nil {
			return fmt.Errorf("failed to send no lesson response during reorder lesson callback handling: %w", err)
		}
		return nil
	}

	jsonedInfo, err := handler.cache.GetInfo(ctx, update.FromChat().ChatConfig().ChatID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info in reorder lesson callback handler: %w", err)
	}
	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info in reorder lesson callback handler: %w", err)
	}

	info.Subject = subject

	infoToStore, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to marshal info to store in database: %w", err)
	}

	err = handler.cache.SaveInfo(ctx, update.FromChat().ChatConfig().ChatID, string(infoToStore))
	if err != nil {
		return fmt.Errorf("failed to save info during reorder lesson callback handler: %w", err)
	}

	_, err = bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.FromChat().ChatConfig().ChatID, info.MarkupMessageId, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove reply markup from message in reorder lesson callback handler: %w", err)
	}
	_, err = bot.SendCtx(ctx, tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID,
		"Хотите ли вы применить данные правила ко всем предметам? Введите \"Да\"/любую другую последовательность символов"))
	if err != nil {
		return fmt.Errorf("failed to send response during reorder lesson callback handler: %w", err)
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.FromChat().ChatConfig().ChatID, constants.REORDER_CHOOSE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during reorder lesson callback handling: %w", err)
	}
	return nil
}

type ReorderConcreteLessonCallbackHandler struct {
	bot   *tgutils.Bot
	cache interfaces.HandlersCache
}

func NewReorderConcreteLessonCallbackHandler(bot *tgutils.Bot, cache interfaces.HandlersCache) *ReorderConcreteLessonCallbackHandler {
	return &ReorderConcreteLessonCallbackHandler{bot: bot, cache: cache}
}

func (handler *ReorderConcreteLessonCallbackHandler) Handle(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	jsonedInfo, err := handler.cache.GetInfo(ctx, update.FromChat().ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info from cache during concrete lesson callback handler: %w", err)
	}
	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info during reorder concrete lesson callback handler: %w", err)
	}

	lessonId, err := parseLessonConcreteCallback(update.CallbackData())
	if err != nil {
		return fmt.Errorf("failed to parse lesson concrete callback data: %w", err)
	}
	info.LessonId = lessonId
	savedInfo, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal info into json during reorder concrete callback: %w", err)
	}
	err = handler.cache.SaveInfo(ctx, update.FromChat().ID, string(savedInfo))
	if err != nil {
		return fmt.Errorf("failed to save info during reorder callback handling: %w", err)
	}

	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(update.FromChat().ID, info.MarkupMessageId, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove reply markup during reorder concrete lesson callback handler: %w", err)
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.FromChat().ID, constants.REORDER_REQUEST_METHOD_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during reorder concrete lesson callback handler: %w", err)
	}
	return nil
}
