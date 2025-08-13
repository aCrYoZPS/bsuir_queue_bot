package labworks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/api/googleapi"
)

var errNoLessons = errors.New("no lessons for given subject")

type SheetsService interface {
	AddLabworkRequest(context.Context, *AppendedLabwork) error
}

type AppendedLabwork struct {
	RequestedDate  time.Time
	SentProofTime  time.Time
	DisciplineName string
	GroupName      string
	FullName       string
	SubgroupNumber int8
	LabworkNumber  int8
}

func NewAppendedLabwork(RequestedDate time.Time, SentProofTime time.Time, DisciplineName string, GroupName string, FullName string, SubgroupNumber, LabworkNumber int8) *AppendedLabwork {
	return &AppendedLabwork{
		RequestedDate:  RequestedDate,
		SentProofTime:  SentProofTime,
		DisciplineName: DisciplineName,
		GroupName:      GroupName,
		FullName:       FullName,
		LabworkNumber:  LabworkNumber,
		SubgroupNumber: SubgroupNumber,
	}
}

type LabworkRequest struct {
	LabworkId      int64     `json:"lab_id,omitempty"`
	RequestedDate  time.Time `json:"requested_time,omitempty"`
	SentProofTime  time.Time `json:"sent_prof,omitempty"`
	DisciplineName string    `json:"discipline,omitempty"`
	GroupName      string    `json:"group,omitempty"`
	SubgroupNumber int8      `json:"subgroup,omitempty"`
	TgId           int64     `json:"tg_id,omitempty"`
	FullName       string    `json:"name,omitempty"`
	LabworkNumber  int8      `json:"lab_num,omitempty"`
	MessageId      int64     `json:"msg_id. omitempty"`
	Notes          string    `json:"notes,omitempty"`
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

func NewLabworksCallbackHandler(bot *tgutils.Bot, cache interfaces.HandlersCache, labworks LabworksService, requests interfaces.RequestsRepository, labworkRequests interfaces.LessonsRequestsRepository, users UsersService, sheets SheetsService) *LabworksCallbackHandler {
	return &LabworksCallbackHandler{
		bot:             bot,
		cache:           cache,
		labworks:        labworks,
		requests:        requests,
		labworkRequests: labworkRequests,
		users:           users,
		sheets:          sheets,
	}
}

func (handler *LabworksCallbackHandler) HandleCallback(ctx context.Context, update *tgbotapi.Update, bot *tgutils.Bot) error {
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
		date, labworkId, subgroup := parseLabworkTimeCallback(update.CallbackData())
		if date.Equal(time.Time{}) || labworkId == 0 {
			return fmt.Errorf("couldn't get time and tg id from labwork callback (%s)", update.CallbackData())
		}
		err := handler.handleTimeCallback(ctx, update.CallbackQuery.Message, date, labworkId, subgroup)
		if err != nil {
			return err
		}
		return nil
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_CONSIDERATION_CALLBACKS) {
		command := strings.TrimPrefix(update.CallbackQuery.Data, constants.LABWORK_CONSIDERATION_CALLBACKS)
		var err error
		switch {
		case strings.HasPrefix(command, "accept"):
			err = handler.handleAcceptCallback(ctx, update.CallbackQuery.Message, command, bot)
		case strings.HasPrefix(command, "decline"):
			err = handler.handleDeclineCallback(ctx, update.CallbackQuery.Message, command, bot)
		default:
			err = fmt.Errorf("no such callback for labworks (%s)", command)
		}
		return err
	}
	if strings.HasPrefix(update.CallbackData(), constants.LABWORK_TIME_CANCEL_CALLBACKS) {
		return handler.handleTimeCancelCallback(ctx, update.CallbackQuery.Message)
	}
	return fmt.Errorf("invalid callback header for labworks callbacks (%s)", update.CallbackData())
}

func (handler *LabworksCallbackHandler) handleDisciplineCallback(ctx context.Context, message *tgbotapi.Message, discipline string) error {
	user, err := handler.users.GetByTgId(ctx, message.Chat.ID)
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
	json, err := json.Marshal(&LabworkRequest{DisciplineName: discipline, GroupName: user.GroupName, FullName: user.FullName, TgId: user.TgId})
	if err != nil {
		return fmt.Errorf("failed to marshal labwork request during labworks discipline callback: %w", err)
	}
	err = handler.cache.SaveInfo(ctx, message.Chat.ID, string(json))
	if err != nil {
		return fmt.Errorf("failed to save info to cache during labworks discipline callback: %w", err)
	}
	keyboard := handler.createDisciplinesKeyboard(lessons)
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, message.MessageID, *keyboard))
	if err != nil {
		return fmt.Errorf("failed to send keyboard during labworks callback handling: %w", err)
	}
	return nil
}

func (handler *LabworksCallbackHandler) createDisciplinesKeyboard(lessons []persistance.Lesson) *tgbotapi.InlineKeyboardMarkup {
	markup := [][]tgbotapi.InlineKeyboardButton{}
	for chunk := range slices.Chunk(lessons, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			formattedDate := fmt.Sprintf("%02d/%02d/%d", discipline.Date.Day(), discipline.Date.Month(), discipline.Date.Year())
			if discipline.SubgroupNumber != iis_api_entities.AllSubgroups {
				formattedDate += fmt.Sprintf(" (%d)", discipline.SubgroupNumber)
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(formattedDate, createLabworkTimeCallback(discipline.Id, discipline.Date, discipline.SubgroupNumber)))
		}
		markup = append(markup, row)
	}
	markup = append(markup, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Назад", constants.LABWORK_TIME_CANCEL_CALLBACKS)))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard
}

func createLabworkTimeCallback(labworkId int64, date time.Time, subgroup iis_api_entities.Subgroup) string {
	builder := strings.Builder{}
	builder.Grow(64)
	builder.WriteString(constants.LABWORK_TIME_CALLBACKS)
	builder.WriteString("|")
	builder.WriteString(fmt.Sprintf("%d.%d.%d", date.Day(), date.Month(), date.Year()))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(labworkId))
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(subgroup))
	return builder.String()
}

func parseLabworkTimeCallback(callback string) (date time.Time, labworkId int64, subgroup int8) {
	callback, _ = strings.CutPrefix(callback, constants.LABWORK_TIME_CALLBACKS+"|")
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
	date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if subgroupString != "" {
		subgroupVal, err := strconv.Atoi(subgroupString)
		if err != nil {
			return time.Time{}, 0, subgroup
		}
		subgroup = int8(subgroupVal)
	}
	return date, labworkId, subgroup
}

func (handler *LabworksCallbackHandler) handleTimeCallback(ctx context.Context, msg *tgbotapi.Message, date time.Time, labworkId int64, subgroup int8) error {
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
	info.SubgroupNumber = subgroup

	infoBytes, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("failed to marshal info during labwork time callback handling: %w", err)
	}
	err = handler.cache.SaveInfo(ctx, msg.Chat.ID, string(infoBytes))
	if err != nil {
		return fmt.Errorf("failed to save info during time callback handling: %w", err)
	}
	err = handler.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.LABWORK_SUBMIT_NUMBER_STATE))
	if err != nil {
		return fmt.Errorf("failed to save labwork submit proof state during labwork time callback handling: %w", err)
	}

	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove markup during labwork time callback handling: %w", err)
	}

	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Введите номер сдаваемой лабораторной работы"))
	if err != nil {
		return fmt.Errorf("failed to send response to user during time callback handling: %w", err)
	}
	return nil
}

func (handler *LabworksCallbackHandler) handleTimeCancelCallback(ctx context.Context, msg *tgbotapi.Message) error {
	markup := [][]tgbotapi.InlineKeyboardButton{{}}
	user, err := handler.users.GetByTgId(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id during handling labwork time cancel callback: %w", err)
	}
	disciplines, err := handler.labworks.GetSubjects(ctx, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get subjects during handling labwork time cancel callback: %w", err)
	}
	for chunk := range slices.Chunk(disciplines, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(discipline, createLabworkDisciplineCallback(msg.From.ID, discipline)))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	_, err = handler.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, keyboard))
	if err != nil {
		return fmt.Errorf("failed to remove markup during labwork time cancel callback handling: %w", err)
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

	err = handler.sheets.AddLabworkRequest(ctx, handler.AppendedLabwork(form))
	if err != nil {
		if googleErr, ok := err.(*googleapi.Error); ok {
			if googleErr.Code == http.StatusInternalServerError {
				_, err := handler.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Ошибка на стороне гугл сервисов. Попробуйте одобрить заявку позже"))
				if err != nil {
					return fmt.Errorf("failed to send google errors failure response during labworks accept callback handling: %w", err)
				}
			}
		}
		return fmt.Errorf("failed to add labwork to sheets during labwork accept callback handling: %w", err)
	}

	err = handler.labworkRequests.Add(ctx, entities.NewLessonRequest(form.LabworkId, form.TgId, int64(msg.MessageID), msg.Chat.ID, form.LabworkNumber))
	if err != nil {
		return fmt.Errorf("failed to add labwork request during labwork accept callback handling: %w", err)
	}

	err = handler.RemoveMarkup(ctx, msg, bot)
	if err != nil {
		return fmt.Errorf("failed to remove markup during labworks accept callback handling: %w", err)
	}

	resp := tgbotapi.NewMessage(form.TgId, "Ваша заявка была принята")
	resp.ReplyToMessageID = int(form.MessageId)
	_, err = bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send message to user during labwork decline callback handling: %w", err)
	}

	return err
}

func (handler *LabworksCallbackHandler) AppendedLabwork(req *LabworkRequest) *AppendedLabwork {
	return &AppendedLabwork{
		RequestedDate:  req.RequestedDate,
		SentProofTime:  req.SentProofTime,
		DisciplineName: req.DisciplineName,
		GroupName:      req.GroupName,
		FullName:       req.FullName,
		SubgroupNumber: req.SubgroupNumber,
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
	resp.ReplyToMessageID = int(form.MessageId)
	_, err = bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send message to user during labwork decline callback handling: %w", err)
	}
	return nil
}
