package labworks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	datetime "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/date_time"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

var (
	errNoLabworks = errors.New("no labworks found")
	errNoDocument = errors.New("no document found")
)

type LabworksService interface {
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
	GetNext(ctx context.Context, subject string, groupId int64) ([]persistance.Lesson, error)
}

type UsersService interface {
	GetByTgId(ctx context.Context, id int64) (*entities.User, error)
}

type labworkSubmitStartState struct {
	bot      *tgutils.Bot
	cache    interfaces.HandlersCache
	labworks LabworksService
	users    UsersService
}

func NewLabworkSubmitStartState(bot *tgutils.Bot, cache interfaces.HandlersCache, labworks LabworksService, users UsersService) *labworkSubmitStartState {
	return &labworkSubmitStartState{bot: bot, cache: cache, labworks: labworks, users: users}
}

func (*labworkSubmitStartState) StateName() string {
	return constants.LABWORK_SUBMIT_START_STATE
}

func (state *labworkSubmitStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(ctx, message.Chat.ID)
	if err != nil {
		return err
	}
	if user.Id == 0 {
		err := state.TransitionAndSend(ctx, tgbotapi.NewMessage(message.Chat.ID, "Вы ещё не присоединились к какой-либо группе"), interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
		return err
	}
	replyMarkup, err := state.createDisciplinesKeyboard(ctx, message.Chat.ID, message.From.ID)
	if err != nil {
		if errors.Is(err, errNoLabworks) {
			return nil
		}
		return err
	}
	resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите предмет и дату пары")
	resp.ReplyMarkup = replyMarkup
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_SUBMIT_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to labwork submit waiting state during labwork submit start state: %w", err)
	}
	sent, err := state.bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send response during labwork submit start state: %w", err)
	}
	json, err := json.Marshal(&LabworkRequest{MarkupMessageId: sent.MessageID})
	if err != nil {
		return fmt.Errorf("failed to marshal labwork request in labwork submit start state: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(json))
	if err != nil {
		return fmt.Errorf("failed to save labwork request in labwork submit start state: %w", err)
	}
	return nil
}

func (state *labworkSubmitStartState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	return nil
}

const CHUNK_SIZE = 4

func (state *labworkSubmitStartState) createDisciplinesKeyboard(ctx context.Context, chatId, userTgId int64) (*tgbotapi.InlineKeyboardMarkup, error) {
	markup := [][]tgbotapi.InlineKeyboardButton{{}}
	user, err := state.users.GetByTgId(ctx, userTgId)
	if err != nil {
		return nil, err
	}
	disciplines, err := state.labworks.GetSubjects(ctx, user.GroupId)
	if err != nil {
		return nil, err
	}
	if len(disciplines) == 0 {
		newState := interfaces.NewCachedInfo(chatId, constants.IDLE_STATE)
		err = state.TransitionAndSend(ctx, tgbotapi.NewMessage(chatId, "Больше не осталось лабораторных. Отдохните"), newState)
		if err != nil {
			return nil, err
		}
		return nil, errNoLabworks
	}
	for chunk := range slices.Chunk(disciplines, CHUNK_SIZE) {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, discipline := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(discipline, createLabworkDisciplineCallback(userTgId, discipline)))
		}
		markup = append(markup, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	return &keyboard, nil
}

func (state *labworkSubmitStartState) TransitionAndSend(ctx context.Context, msg tgbotapi.MessageConfig, newState *interfaces.CachedInfo) error {
	err := state.cache.SaveState(ctx, *newState)
	if err != nil {
		return err
	}
	_, err = state.bot.SendCtx(ctx, msg)
	return err
}

func createLabworkDisciplineCallback(userTgId int64, discipline string) string {
	builder := strings.Builder{}
	builder.Grow(64)
	builder.WriteString(constants.LABWORK_DISCIPLINE_CALLBACKS)
	builder.WriteString("|")
	builder.WriteString(discipline)
	builder.WriteString("|")
	builder.WriteString(fmt.Sprint(userTgId))
	return builder.String()
}

func parseLabworkDisciplineCallback(callback string) (discipline string, userTgId int64) {
	callback, _ = strings.CutPrefix(callback, constants.LABWORK_DISCIPLINE_CALLBACKS+"|")
	discipline, after, _ := strings.Cut(callback, "|")
	userTgId, err := strconv.ParseInt(after, 10, 64)
	if err != nil {
		return "", 0
	}
	return discipline, userTgId
}

type StateMachine interface {
	HandleState(ctx context.Context, msg *tgbotapi.Message) error
}
type labworkSubmitWaitingState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	machine StateMachine
}

func NewLabworkSubmitWaitingState(bot *tgutils.Bot, cache interfaces.HandlersCache, machine StateMachine) *labworkSubmitWaitingState {
	return &labworkSubmitWaitingState{bot: bot, cache: cache, machine: machine}
}

func (*labworkSubmitWaitingState) StateName() string {
	return constants.LABWORK_SUBMIT_WAITING_STATE
}

func (state *labworkSubmitWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, закончите отправление заявки, прежде чем переходить к остальным командам"))
	if err != nil {
		return fmt.Errorf("couldn't send wait message: %v", err)
	}
	return nil
}

func (state *labworkSubmitWaitingState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	info, err := state.cache.GetInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info from cache during reverting labwork submit waiting state: %w", err)
	}
	request := LabworkRequest{}
	err = json.Unmarshal([]byte(info), &request)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info into labwork request in %s: %w", state.StateName(), err)
	}

	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, request.MarkupMessageId,
		tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove markup during %s from message in chat with %d: %w", state.StateName(), msg.Chat.ID, err)
	}
	err = state.cache.RemoveInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove cache in %s: %w", state.StateName(), err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during reverting labwork submit waiting state: %w", err)
	}
	return nil
}

type GroupsService interface {
	GetAdmins(ctx context.Context, groupName string) ([]entities.User, error)
}

type labworkSubmitNumberState struct {
	bot          *tgutils.Bot
	cache        interfaces.HandlersCache
	users        UsersService
	labworks     LabworksService
	stateMachine StateMachine
}

func NewLabworkSubmitNumberState(bot *tgutils.Bot, cache interfaces.HandlersCache, labworks LabworksService, users UsersService, stateMachine StateMachine) *labworkSubmitNumberState {
	return &labworkSubmitNumberState{bot: bot, cache: cache, labworks: labworks, users: users, stateMachine: stateMachine}
}

func (*labworkSubmitNumberState) StateName() string {
	return constants.LABWORK_SUBMIT_NUMBER_STATE
}

func (state *labworkSubmitNumberState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	num, err := strconv.ParseUint(message.Text, 10, 8)
	if err != nil || num == 0 || num > 255 {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста,введите корректный номер лабораторной (одно число, в разумных пределах)"))
		if err != nil {
			return fmt.Errorf("failed to send incorrect number msg during labwork submit number state: %w", err)
		}
		return nil
	}
	jsonString, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return err
	}
	req := LabworkRequest{}
	err = json.Unmarshal([]byte(jsonString), &req)
	if err != nil {
		return err
	}
	req.LabworkNumber = int8(num)
	//If it could be correctly unmarshalled, it could be correctly marshalled
	bytes, _ := json.Marshal(&req)
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(bytes))
	if err != nil {
		return fmt.Errorf("failed to save info during labwork submit name state: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_SUBMIT_PROOF_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to labwork proof submit state during labwork submit number state handling: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите доказательство готовности лабораторной работы (один прикрепленный файл, возможно с текстовой подписью)"))
	if err != nil {
		return fmt.Errorf("failed to send response during labwork number submit state: %w", err)
	}
	return nil
}

func (state *labworkSubmitNumberState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.LABWORK_SUBMIT_START_STATE))
	if err != nil {
		return fmt.Errorf("failed to save %s state during labwork submit number state reversion: %w, err", constants.LABWORK_SUBMIT_NUMBER_STATE, err)
	}
	msg.Text = "/submit"
	err = state.stateMachine.HandleState(ctx, msg)
	return err
}

type labworkSubmitProofState struct {
	bot      *tgutils.Bot
	cache    interfaces.HandlersCache
	groups   GroupsService
	requests interfaces.RequestsRepository
}

func NewLabworkSubmitProofState(bot *tgutils.Bot, cache interfaces.HandlersCache, groups GroupsService, requests interfaces.RequestsRepository) *labworkSubmitProofState {
	return &labworkSubmitProofState{bot: bot, cache: cache, groups: groups, requests: requests}
}

func (*labworkSubmitProofState) StateName() string {
	return constants.LABWORK_SUBMIT_PROOF_STATE
}

func (state *labworkSubmitProofState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	jsonString, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return err
	}
	req := &LabworkRequest{}
	err = json.Unmarshal([]byte(jsonString), req)
	if err != nil {
		return err
	}
	req.SentProofTime = datetime.DateTime(time.Now())
	req.MessageId = int64(message.MessageID)

	jsonedReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedReq))
	if err != nil {
		return err
	}
	admins, err := state.groups.GetAdmins(ctx, req.GroupName)
	if err != nil {
		return err
	}
	err = state.handleDocumentType(ctx, admins, message, req)
	if err != nil {
		if !errors.Is(err, errNoDocument) {
			return err
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "")
		req.Notes = message.Text
		err = state.SendMessagesToAdmins(ctx, admins, &msg, req)
		if err != nil {
			return fmt.Errorf("failed to send messages to admins during labwork proof submit state: %w", err)
		}
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return err
	}
	return err
}

func (state *labworkSubmitProofState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.LABWORK_SUBMIT_NUMBER_STATE))
	if err != nil {
		return fmt.Errorf("failed to change state while reverting labwork submit proof state: %w", err)
	}
	return nil
}

func (state *labworkSubmitProofState) handleDocumentType(ctx context.Context, admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	form.Notes = message.Caption
	var err error
	switch {
	case message.Photo != nil:
		err = state.handlePhotoProof(ctx, admins, message, form)
	case message.Document != nil:
		err = state.handleDocumentProof(ctx, admins, message, form)
	default:
		return errNoDocument
	}
	if errors.Is(err, tgutils.ErrMsgInvalidLen) {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Извините,ваше сообщение слишком большое для отправки. Измените его и отправьте снова"))
		if err != nil {
			return fmt.Errorf("failed to send too large message during labwork submit proof state: %w", err)
		}
		return nil
	}
	return err
}

func (state *labworkSubmitProofState) handlePhotoProof(ctx context.Context, admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	maxSizeId := tgutils.SelectMaxSizedPhoto(message.Photo)
	fileBytes, err := state.GetFileBytes(maxSizeId)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "rnd_name", Bytes: fileBytes})
	err = state.SendPhotosToAdmins(ctx, admins, &msg, form)
	return err
}

func (state *labworkSubmitProofState) handleDocumentProof(ctx context.Context, admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	maxSizeId := message.Document.FileID
	fileBytes, err := state.GetFileBytes(maxSizeId)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FileBytes{Name: message.Document.FileName, Bytes: fileBytes})
	err = state.SendDocumentsToAdmins(ctx, admins, &msg, form)
	return err
}

func (state *labworkSubmitProofState) GetFileBytes(fileId string) ([]byte, error) {
	file, err := state.bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		return nil, err
	}
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, errors.New("couldn't receive bot_token")
	}
	link := file.Link(os.Getenv("BOT_TOKEN"))
	resp, err := http.Get(link)
	if err != nil && err != io.EOF {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

var funcMap = template.FuncMap{"dateTime": func(ts datetime.DateTime) string {
	t := time.Time(ts)
	return fmt.Sprintf("%02d.%02d.%02d %02d:%02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second())
}, "date": func(dt datetime.DateTime) string {
	t := time.Time(dt)
	return fmt.Sprintf("%02d.%02d.%d", t.Day(), t.Month(), t.Year())
}}

const tmplText = "Отправил: {{.FullName}}\nПредмет: {{.DisciplineName}}\nНомер лабораторной: {{.LabworkNumber}}\nДата: {{date .RequestedDate}}\nВремя отправки: {{dateTime .SentProofTime}}\n{{if .Notes}}Доп информация: {{.Notes}} {{end}}"

var adminSendingTmpl = template.Must(template.New("adminProofSent").Funcs(funcMap).Parse(tmplText))

func (state *labworkSubmitProofState) SendPhotosToAdmins(ctx context.Context, admins []entities.User, photo *tgbotapi.PhotoConfig, form *LabworkRequest) error {
	var buf bytes.Buffer
	err := adminSendingTmpl.Execute(&buf, form)
	if err != nil {
		return err
	}
	photo.ReplyMarkup = createMarkupKeyboard(form)
	photo.Caption = buf.String()
	reqUUID := uuid.NewString()
	for _, admin := range admins {
		photo.ChatID = admin.TgId
		sentMsg, err := state.bot.SendCtx(ctx, photo)
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(ctx, interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (state *labworkSubmitProofState) SendMessagesToAdmins(ctx context.Context, admins []entities.User, msg *tgbotapi.MessageConfig, form *LabworkRequest) error {
	var buf bytes.Buffer
	err := adminSendingTmpl.Execute(&buf, form)
	if err != nil {
		return err
	}
	msg.ReplyMarkup = createMarkupKeyboard(form)
	msg.Text = buf.String()
	reqUUID := uuid.NewString()
	for _, admin := range admins {
		msg.ChatID = admin.TgId
		sentMsg, err := state.bot.SendCtx(ctx, msg)
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(ctx, interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (state *labworkSubmitProofState) SendDocumentsToAdmins(ctx context.Context, admins []entities.User, msg *tgbotapi.DocumentConfig, form *LabworkRequest) error {
	var buf bytes.Buffer
	err := adminSendingTmpl.Execute(&buf, form)
	if err != nil {
		return err
	}
	msg.ReplyMarkup = createMarkupKeyboard(form)
	msg.Caption = buf.String()
	reqUUID := uuid.NewString()
	for _, admin := range admins {
		msg.ChatID = admin.TgId
		sentMsg, err := state.bot.SendCtx(ctx, msg)
		if err != nil {
			return fmt.Errorf("couldn't send documents to admins as proof: %v", err)
		}
		err = state.requests.SaveRequest(ctx, interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
		if err != nil {
			return fmt.Errorf("couldn't send documents to admins as proof: %v", err)
		}
	}
	return nil
}

func createMarkupKeyboard(form *LabworkRequest) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.LABWORK_CONSIDERATION_CALLBACKS + "accept" + fmt.Sprint(form.TgId)
	declineData := constants.LABWORK_CONSIDERATION_CALLBACKS + "decline" + fmt.Sprint(form.TgId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Принять", acceptData), tgbotapi.NewInlineKeyboardButtonData("Отклонить", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
