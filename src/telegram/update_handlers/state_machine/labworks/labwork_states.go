package labworks

import (
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
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

var errNoLabworks = errors.New("no labworks found")

type LabworksService interface {
	GetSubjects(groupId int64) ([]string, error)
	GetNext(subject string, groupId int64) ([]persistance.Lesson, error)
}

type UsersService interface {
	GetByTgId(id int64) (*entities.User, error)
}

type labworkSubmitStartState struct {
	bot      *tgbotapi.BotAPI
	cache    interfaces.HandlersCache
	labworks LabworksService
	users    UsersService
}

func NewLabworkSubmitStartState(bot *tgbotapi.BotAPI, cache interfaces.HandlersCache, labworks LabworksService, users UsersService) *labworkSubmitStartState {
	return &labworkSubmitStartState{bot: bot, cache: cache, labworks: labworks, users: users}
}

func (*labworkSubmitStartState) StateName() string {
	return constants.LABWORK_SUBMIT_START_STATE
}

func (state *labworkSubmitStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(message.From.ID)
	if err != nil {
		return err
	}
	if user == nil {
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Вы ещё не присоединились к какой-либо группе"))
		return err
	}
	replyMarkup, err := state.createDisciplinesKeyboard(message.Chat.ID, message.From.ID)
	if err != nil {
		if errors.Is(err, errNoLabworks) {
			return nil
		}
		return err
	}
	resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите предмет и дату пары")
	resp.ReplyMarkup = replyMarkup
	err = state.TransitionAndSend(resp, interfaces.NewCachedInfo(message.Chat.ID, constants.LABWORK_SUBMIT_WAITING_STATE))
	return err
}

const CHUNK_SIZE = 4

func (state *labworkSubmitStartState) createDisciplinesKeyboard(chatId, userTgId int64) (*tgbotapi.InlineKeyboardMarkup, error) {
	markup := [][]tgbotapi.InlineKeyboardButton{{}}
	user, err := state.users.GetByTgId(userTgId)
	if err != nil {
		return nil, err
	}
	disciplines, err := state.labworks.GetSubjects(user.GroupId)
	if err != nil {
		return nil, err
	}
	if len(disciplines) == 0 {
		newState := interfaces.NewCachedInfo(chatId, constants.IDLE_STATE)
		err = state.TransitionAndSend(tgbotapi.NewMessage(chatId, "Больше не осталось лабораторных. Отдохните"), newState)
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

func (state *labworkSubmitStartState) TransitionAndSend(msg tgbotapi.MessageConfig, newState *interfaces.CachedInfo) error {
	err := state.cache.SaveState(*newState)
	if err != nil {
		return err
	}
	_, err = state.bot.Send(msg)
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

type labworkSubmitWaitingState struct {
	bot *tgbotapi.BotAPI
}

func NewLabworkSubmitWaitingState(bot *tgbotapi.BotAPI) *labworkSubmitWaitingState {
	return &labworkSubmitWaitingState{bot: bot}
}

func (*labworkSubmitWaitingState) StateName() string {
	return constants.LABWORK_SUBMIT_WAITING_STATE
}

func (state *labworkSubmitWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	errRes := make(chan error, 1)
	go func(chan error) {
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, закончите отправление заявки, прежде чем переходить к остальным командам"))
		errRes <- err
	}(errRes)
	select {
	case <-ctx.Done():
		return fmt.Errorf("couldn't send labwork waiting message: %v", ctx.Err())
	case err := <-errRes:
		return err
	}
}

type GroupsService interface {
	GetAdmins(groupName string) ([]entities.User, error)
}

type labworkSubmitProofState struct {
	bot      *tgbotapi.BotAPI
	cache    interfaces.HandlersCache
	groups   GroupsService
	requests interfaces.RequestsRepository
}

func NewLabworkSubmitProofState(bot *tgbotapi.BotAPI, cache interfaces.HandlersCache, groups GroupsService, requests interfaces.RequestsRepository) *labworkSubmitProofState {
	return &labworkSubmitProofState{bot: bot, cache: cache, groups: groups, requests: requests}
}

func (*labworkSubmitProofState) StateName() string {
	return constants.LABWORK_SUBMIT_PROOF_STATE
}

func (state *labworkSubmitProofState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	caption := message.Caption
	before, _, _ := strings.Cut(caption, " ")
	labworkNum, err := strconv.Atoi(before)
	if err != nil {
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введите корректный номер группы"))
		return err
	}
	jsonString, err := state.cache.GetInfo(message.Chat.ID)
	if err != nil {
		return err
	}
	req := &LabworkRequest{}
	err = json.Unmarshal([]byte(jsonString), req)
	if err != nil {
		return err
	}
	req.LabworkNumber = int8(labworkNum)
	req.SentProofTime = time.Now()

	jsonedReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(message.Chat.ID, string(jsonedReq))
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return err
	}
	admins, err := state.groups.GetAdmins(req.GroupName)
	if err != nil {
		return err
	}
	err = state.handleDocumentType(ctx, admins, message, req)
	return err
}

func (state *labworkSubmitProofState) handleDocumentType(ctx context.Context, admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	var err error
	switch {
	case message.Photo != nil:
		err = state.handlePhotoProof(admins, message, form)
	case message.Document != nil:
		err = state.handleDocumentProof(ctx, admins, message, form)
	default:
		err = errors.New("no document in a given message")
	}
	return err
}

func (state *labworkSubmitProofState) handlePhotoProof(admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	maxSizeId := tgutils.SelectMaxSizedPhoto(message.Photo)
	fileBytes, err := state.GetFileBytes(maxSizeId)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "rnd_name", Bytes: fileBytes})
	state.SendPhotosToAdmins(admins, &msg, form)
	return nil
}

func (state *labworkSubmitProofState) handleDocumentProof(ctx context.Context, admins []entities.User, message *tgbotapi.Message, form *LabworkRequest) error {
	maxSizeId := tgutils.SelectMaxSizedPhoto(message.Photo)
	fileBytes, err := state.GetFileBytes(maxSizeId)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FileBytes{Name: "rnd_name", Bytes: fileBytes})
	state.SendDocumentsToAdmins(ctx, admins, &msg, form)
	return nil
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

const adminSendingTmpl = "Предмет %s\nЛабораторная:%sДата:%sОтправил: %s\n"

func (state *labworkSubmitProofState) SendPhotosToAdmins(admins []entities.User, photo *tgbotapi.PhotoConfig, form *LabworkRequest) error {
	text := fmt.Sprintf(adminSendingTmpl, form.DisciplineName, fmt.Sprint(form.LabworkNumber),
		fmt.Sprintf("%d.%d.%d", form.RequestedDate.Day(), form.RequestedDate.Month(), form.RequestedDate.Year()), form.FullName)
	photo.ReplyMarkup = createMarkupKeyboard(form)
	photo.Caption = text
	reqUUID := uuid.NewString()
	for _, admin := range admins {
		photo.ChatID = admin.TgId
		sentMsg, err := state.bot.Send(photo)
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (state *labworkSubmitProofState) SendDocumentsToAdmins(ctx context.Context, admins []entities.User, msg *tgbotapi.DocumentConfig, form *LabworkRequest) error {
	errRes := make(chan error, 1)
	go func(chan error) {
		text := fmt.Sprintf(adminSendingTmpl, form.DisciplineName, fmt.Sprint(form.LabworkNumber),
			fmt.Sprintf("%d.%d.%d", form.RequestedDate.Day(), form.RequestedDate.Month(), form.RequestedDate.Year()), form.FullName)
		msg.ReplyMarkup = createMarkupKeyboard(form)
		msg.Caption = text
		reqUUID := uuid.NewString()
		for _, admin := range admins {
			msg.ChatID = admin.TgId
			sentMsg, err := state.bot.Send(msg)
			if err != nil {
				errRes <- err
			}
			err = state.requests.SaveRequest(interfaces.NewGroupRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, interfaces.WithUUID(reqUUID)))
			if err != nil {
				errRes <- err
			}
		}
		errRes <- nil
	}(errRes)
	select {
	case <-ctx.Done():
		return fmt.Errorf("couldn't send documents to admins as proof: %v", ctx.Err())
	case err := <-errRes:
		return err
	}
}

func createMarkupKeyboard(form *LabworkRequest) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.LABWORK_CONSIDERATION_CALLBACKS + "accept" + fmt.Sprint(form.TgId)
	declineData := constants.LABWORK_CONSIDERATION_CALLBACKS + "decline" + fmt.Sprint(form.TgId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Принять", acceptData), tgbotapi.NewInlineKeyboardButtonData("Отклонить", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
