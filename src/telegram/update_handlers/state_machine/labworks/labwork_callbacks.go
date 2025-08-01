package labworks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var errNoLessons = errors.New("no lessons for given subject")

type SheetsService interface {
	AddLabwork(context.Context, *LabworkRequest) error
}

type LabworkRequest struct {
	RequestedDate  time.Time `json:"requested_time,omitempty"`
	SentProofTime  time.Time `json:"sent_prof,omitempty"`
	DisciplineName string    `json:"discipline,omitempty"`
	GroupName      string    `json:"group,omitempty"`
	TgId           int64     `json:"tg_id,omitempty"`
	FullName       string    `json:"name,omitempty"`
	LabworkNumber  int8      `json:"lab_num,omitempty"`
}

type LabworksCallbackHandler struct {
	bot      *tgbotapi.BotAPI
	cache    interfaces.HandlersCache
	requests interfaces.RequestsRepository
	sheets   SheetsService
	labworks LabworksService
	users    UsersService
}

func NewLabworksCallbackHandler(bot *tgbotapi.BotAPI, cache interfaces.HandlersCache, labworks LabworksService, users UsersService, sheets SheetsService) *LabworksCallbackHandler {
	return &LabworksCallbackHandler{
		bot:      bot,
		cache:    cache,
		labworks: labworks,
		users:    users,
		sheets:   sheets,
	}
}

func (handler *LabworksCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return err
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_DISCIPLINE_CALLBACKS) {
		discipline, userTgId := parseLabworkDisciplineCallback(update.CallbackData())
		if discipline == "" || userTgId == 0 {
			return errors.New("invalid command requested")
		}
		err := handler.handleDisciplineCallback(ctx, update.CallbackQuery.Message, discipline)
		if err != nil {
			if errors.Is(err, errNoLessons) {
				err := handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(update.CallbackQuery.Message.Chat.ID, constants.IDLE_STATE))
				if err != nil {
					return err
				}
				_, err = bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, cases.Title(language.English).String(errNoLessons.Error())))
				return err
			}
			return err
		}
		return nil
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_TIME_CALLBACKS) {
		date, userTgId := parseLabworkTimeCallback(update.CallbackData())
		if date.Equal(time.Time{}) || userTgId == 0 {
			return errors.New("invalid callback values")
		}
		err := handler.handleTimeCallback(ctx, update.CallbackQuery.Message, date)
		if err != nil {
			return err
		}
		return nil
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_CONSIDERATION_CALLBACKS) {
		command := strings.TrimPrefix(update.CallbackQuery.Data, constants.ADMIN_CALLBACKS)
		var err error
		switch {
		case strings.HasPrefix(command, "accept"):
			err = handler.handleAcceptCallback(ctx, update.CallbackQuery.Message, command, bot)
		case strings.HasPrefix(command, "decline"):
			err = handler.handleDeclineCallback(ctx, update.CallbackQuery.Message, command, bot)
		default:
			err = errors.New("no such callback")
		}
		if err != nil {
			return err
		}
	}
	return errors.New("invalid callback header")
}

func (handler *LabworksCallbackHandler) handleDisciplineCallback(ctx context.Context, message *tgbotapi.Message, discipline string) error {
	user, err := handler.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return err
	}
	lessons, err := handler.labworks.GetNext(ctx, discipline, user.GroupId)
	if err != nil {
		return err
	}
	if lessons == nil {
		return errNoLessons
	}
	json, err := json.Marshal(&LabworkRequest{DisciplineName: discipline, GroupName: user.GroupName, FullName: user.GroupName, TgId: user.TgId})
	if err != nil {
		return err
	}
	err = handler.cache.SaveInfo(ctx, message.Chat.ID, string(json))
	if err != nil {
		return err
	}
	keyboard, err := handler.createDisciplinesKeyboard(lessons, message.From.ID)
	if err != nil {
		return err
	}
	_, err = handler.bot.Send(tgbotapi.NewEditMessageReplyMarkup(message.From.ID, message.MessageID, *keyboard))
	return err
}

func (handler *LabworksCallbackHandler) createDisciplinesKeyboard(lessons []persistance.Lesson, userTgId int64) (*tgbotapi.InlineKeyboardMarkup, error) {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(lessons, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			formattedDate := fmt.Sprintf("%d.%d.%d", discipline.Date.Day(), discipline.Date.Month(), discipline.Date.Year())
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(formattedDate, createLabworkTimeCallback(userTgId, discipline.Date)))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard, nil
}

func createLabworkTimeCallback(userTgId int64, date time.Time) string {
	builder := strings.Builder{}
	builder.Grow(64)
	builder.WriteString(constants.LABWORK_TIME_CALLBACKS)
	builder.WriteString("|")
	builder.WriteString(fmt.Sprintf("%d.%d.%d", date.Day(), date.Month(), date.Year()))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(userTgId))
	return builder.String()
}

func parseLabworkTimeCallback(callback string) (date time.Time, userTgId int64) {
	callback, _ = strings.CutPrefix(callback, constants.LABWORK_TIME_CALLBACKS+"|")
	formattedDate, after, _ := strings.Cut(callback, "|")
	userTgId, err := strconv.ParseInt(after, 10, 64)
	if err != nil {
		return time.Time{}, 0
	}
	nums := strings.Split(formattedDate, ".")
	if len(nums) < 3 {
		return time.Time{}, 0
	}
	day, err := strconv.Atoi(nums[0])
	if err != nil {
		return time.Time{}, 0
	}
	month, err := strconv.Atoi(nums[1])
	if err != nil {
		return time.Time{}, 0
	}
	year, err := strconv.Atoi(nums[2])
	if err != nil {
		return time.Time{}, 0
	}
	date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date, userTgId
}

func (handler *LabworksCallbackHandler) handleTimeCallback(ctx context.Context, msg *tgbotapi.Message, date time.Time) error {
	jsonedInfo, err := handler.cache.GetInfo(ctx, msg.Chat.ID)
	if err != nil {
		return err
	}
	info := LabworkRequest{}
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return err
	}
	info.RequestedDate = date
	infoBytes, err := json.Marshal(&info)
	if err != nil {
		return err
	}
	err = handler.cache.SaveInfo(ctx, msg.Chat.ID, string(infoBytes))
	if err != nil {
		return err
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.LABWORK_SUBMIT_PROOF_STATE))
	if err != nil {
		return err
	}
	_, err = handler.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите доказательство готовности лабораторной работы (один прикрепленный файл) и её номер (формат подписи: 4 лаба сделана)"))
	return err
}

func (handler *LabworksCallbackHandler) handleAcceptCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &LabworkRequest{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}

	err = handler.sheets.AddLabwork(ctx, form)
	if err != nil {
		return err
	}

	err = handler.RemoveMarkup(ctx, msg, bot)
	return err
}

func (handler *LabworksCallbackHandler) handleDeclineCallback(ctx context.Context,msg *tgbotapi.Message, command string, bot *tgbotapi.BotAPI) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return err
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return err
	}
	form := &LabworkRequest{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return err
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return err
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return err
	}
	resp := tgbotapi.NewMessage(form.TgId, "Ваша заявка была отклонена")
	_, err = bot.Send(resp)
	return err
}

func (handler *LabworksCallbackHandler) RemoveMarkup(ctx context.Context, msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	request, err := handler.requests.GetByMsg(ctx, int64(msg.MessageID), msg.Chat.ID)
	if err != nil {
		return err
	}
	requests, err := handler.requests.GetByUUID(ctx, request.UUID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		err = handler.requests.DeleteRequest(ctx, request.MsgId)
		if err != nil {
			return err
		}
		_, err := bot.Send(tgbotapi.NewEditMessageReplyMarkup(request.ChatId, int(request.MsgId), tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return err
		}
	}
	return nil
}
