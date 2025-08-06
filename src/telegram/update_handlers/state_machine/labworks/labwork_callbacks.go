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

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var errNoLessons = errors.New("no lessons for given subject")

type SheetsService interface {
	AddLabwork(context.Context, *AppendedLabwork) error
}

type AppendedLabwork struct {
	RequestedDate  time.Time
	SentProofTime  time.Time
	DisciplineName string
	GroupName      string
	FullName       string
	LabworkNumber  int8
}

func NewAppendedLabwork(RequestedDate time.Time, SentProofTime time.Time, DisciplineName string, GroupName string, FullName string, LabworkNumber int8) *AppendedLabwork {
	return &AppendedLabwork{
		RequestedDate:  RequestedDate,
		SentProofTime:  SentProofTime,
		DisciplineName: DisciplineName,
		GroupName:      GroupName,
		FullName:       FullName,
		LabworkNumber:  LabworkNumber,
	}
}

type LabworkRequest struct {
	LabworkId      int64     `json:"lab_id,omitempty"`
	RequestedDate  time.Time `json:"requested_time,omitempty"`
	SentProofTime  time.Time `json:"sent_prof,omitempty"`
	DisciplineName string    `json:"discipline,omitempty"`
	GroupName      string    `json:"group,omitempty"`
	TgId           int64     `json:"tg_id,omitempty"`
	FullName       string    `json:"name,omitempty"`
	LabworkNumber  int8      `json:"lab_num,omitempty"`
}

type LabworksCallbackHandler struct {
	bot             *tgutils.Bot
	cache           interfaces.HandlersCache
	requests        interfaces.RequestsRepository
	labworkRequests interfaces.LessonsRequestsRepository
	sheets          SheetsService
	labworks        LabworksService
	users           UsersService
}

func NewLabworksCallbackHandler(bot *tgutils.Bot, cache interfaces.HandlersCache, labworks LabworksService, labworkRequests interfaces.LessonsRequestsRepository, users UsersService, sheets SheetsService) *LabworksCallbackHandler {
	return &LabworksCallbackHandler{
		bot:             bot,
		cache:           cache,
		labworks:        labworks,
		labworkRequests: labworkRequests,
		users:           users,
		sheets:          sheets,
	}
}

func (handler *LabworksCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		return fmt.Errorf("failed to create labwork callback while handling: %w", err)
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
					return fmt.Errorf("failed to transition to idle state during labwork callback handling: %w", err)
				}
				_, err = bot.SendCtx(ctx, tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, cases.Title(language.English).String(errNoLessons.Error())))
				return fmt.Errorf("failed to send no lessons error to user during labwork callback handling: %w", err)
			}
			return err
		}
		return nil
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_TIME_CALLBACKS) {
		date, labworkId := parseLabworkTimeCallback(update.CallbackData())
		if date.Equal(time.Time{}) || labworkId == 0 {
			return fmt.Errorf("couldn't get time and tg id from labwork callback (%s)", update.CallbackData())
		}
		err := handler.handleTimeCallback(ctx, update.CallbackQuery.Message, date, labworkId)
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
			err = fmt.Errorf("no such callback for labworks (%s)", command)
		}
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("invalid callback header for labworks callbacks (%s)", update.CallbackData())
}

func (handler *LabworksCallbackHandler) handleDisciplineCallback(ctx context.Context, message *tgbotapi.Message, discipline string) error {
	user, err := handler.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during labworks discipline callback: %w", err)
	}
	lessons, err := handler.labworks.GetNext(ctx, discipline, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get next labworks during labworks discipline callback: %w", err)
	}
	if lessons == nil {
		return errNoLessons
	}
	json, err := json.Marshal(&LabworkRequest{DisciplineName: discipline, GroupName: user.GroupName, FullName: user.GroupName, TgId: user.TgId})
	if err != nil {
		return fmt.Errorf("failed to marshal labwork request during labworks discipline callback: %w", err)
	}
	err = handler.cache.SaveInfo(ctx, message.Chat.ID, string(json))
	if err != nil {
		return fmt.Errorf("failed to save info to cache during labworks discipline callback: %w", err)
	}
	keyboard := handler.createDisciplinesKeyboard(lessons, message.From.ID)
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(message.From.ID, message.MessageID, *keyboard))
	if err != nil {
		return fmt.Errorf("failed to send keyboard during labworks callback handling: %w", err)
	}
	return nil
}

func (handler *LabworksCallbackHandler) createDisciplinesKeyboard(lessons []persistance.Lesson, userTgId int64) *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(lessons, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			formattedDate := fmt.Sprintf("%d.%d.%d", discipline.Date.Day(), discipline.Date.Month(), discipline.Date.Year())
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(formattedDate, createLabworkTimeCallback(discipline.Id, discipline.Date)))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}

func createLabworkTimeCallback(labworkId int64, date time.Time) string {
	builder := strings.Builder{}
	builder.Grow(64)
	builder.WriteString(constants.LABWORK_TIME_CALLBACKS)
	builder.WriteString("|")
	builder.WriteString(fmt.Sprintf("%d.%d.%d", date.Day(), date.Month(), date.Year()))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(labworkId))
	return builder.String()
}

func parseLabworkTimeCallback(callback string) (date time.Time, labworkId int64) {
	callback, _ = strings.CutPrefix(callback, constants.LABWORK_TIME_CALLBACKS+"|")
	formattedDate, after, _ := strings.Cut(callback, "|")
	labworkId, err := strconv.ParseInt(after, 10, 64)
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
	return date, labworkId
}

func (handler *LabworksCallbackHandler) handleTimeCallback(ctx context.Context, msg *tgbotapi.Message, date time.Time, labworkId int64) error {
	jsonedInfo, err := handler.cache.GetInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info during labwork time callback handling: %w", err)
	}
	info := LabworkRequest{}
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info during labwork time callback handling: %w", err)
	}

	info.RequestedDate = date
	info.LabworkId = labworkId

	infoBytes, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to marshal info during labwork time callback handling: %w", err)
	}
	err = handler.cache.SaveInfo(ctx, msg.Chat.ID, string(infoBytes))
	if err != nil {
		return fmt.Errorf("failed to save info during time callback handling: %w", err)
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.LABWORK_SUBMIT_PROOF_STATE))
	if err != nil {
		return fmt.Errorf("failed to save labwork submit proof state during labwork time callback handling: %w", err)
	}
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Введите доказательство готовности лабораторной работы (один прикрепленный файл) и её номер (формат подписи: 4 лаба сделана)"))
	if err != nil {
		return fmt.Errorf("failed to send response to user during time callback handling: %w", err)
	}
	return nil
}

func (handler *LabworksCallbackHandler) handleAcceptCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	chatId, err := strconv.ParseInt(strings.TrimPrefix(command, "accept"), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to get chat id from command (%s) during labwork accept callback handling: %w", strings.TrimPrefix(command, "accept"), err)
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return fmt.Errorf("failed to get info during labwork accept callback handling: %w", err)
	}
	form := &LabworkRequest{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info during labwork accept callback handling: %w", err)
	}

	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during labwork accept callback handling: %w", err)
	}

	err = handler.sheets.AddLabwork(ctx, handler.AppendedLabwork(form))
	if err != nil {
		return fmt.Errorf("failed to add labwork to sheets during labwork accept callback handling: %w", err)
	}

	err = handler.labworkRequests.Add(ctx, entities.NewLessonRequest(form.LabworkId, form.TgId, int64(msg.MessageID), msg.Chat.ID, form.LabworkNumber))
	if err != nil {
		return fmt.Errorf("failed to add labwork request during labwork accept callback handling: %w", err)
	}

	err = handler.RemoveMarkup(ctx, msg, bot)
	return err
}

func (handler *LabworksCallbackHandler) AppendedLabwork(req *LabworkRequest) *AppendedLabwork {
	return &AppendedLabwork{
		RequestedDate:  req.RequestedDate,
		SentProofTime:  req.SentProofTime,
		DisciplineName: req.DisciplineName,
		GroupName:      req.DisciplineName,
		FullName:       req.FullName,
		LabworkNumber:  req.LabworkNumber,
	}
}

func (handler *LabworksCallbackHandler) RemoveMarkup(ctx context.Context, msg *tgbotapi.Message, bot *tgutils.Bot) error {
	request, err := handler.requests.GetByMsg(ctx, int64(msg.MessageID), msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get labwork request by message during markup removal: %w", err)
	}
	requests, err := handler.requests.GetByUUID(ctx, request.UUID)
	if err != nil {
		return fmt.Errorf("failed to get labwork requests by uuid during markup removal: %w", err)
	}
	for _, request := range requests {
		err = handler.requests.DeleteRequest(ctx, request.MsgId)
		if err != nil {
			return fmt.Errorf("failed to delete labwork request during markup removal: %w", err)
		}
		_, err := bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(request.ChatId, int(request.MsgId), tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
		if err != nil {
			return fmt.Errorf("failed to send response message during markup removal: %w", err)
		}
	}
	return nil
}

func (handler *LabworksCallbackHandler) handleDeclineCallback(ctx context.Context, msg *tgbotapi.Message, command string, bot *tgutils.Bot) error {
	var chatId int64
	err := json.Unmarshal([]byte(strings.TrimPrefix(command, "decline")), &chatId)
	if err != nil {
		return fmt.Errorf("failed to unmarshal chat id from command (%s) during labwork decline callback handling: %w", command, err)
	}
	info, err := handler.cache.GetInfo(ctx, chatId)
	if err != nil {
		return fmt.Errorf("failed to get info from cache during labwork decline callback handling: %w", err)
	}
	form := &LabworkRequest{}
	err = json.Unmarshal([]byte(info), &form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info (%s) during labwork decline callback handling: %w", info, err)
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(chatId, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during labwork decline callback handling: %w", err)
	}
	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return err
	}
	resp := tgbotapi.NewMessage(form.TgId, "Ваша заявка была отклонена")
	_, err = bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send message to user during labwork decline callback handling: %w", err)
	}
	return nil
}
