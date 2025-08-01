package admin

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

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/constants"
	stateErrors "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/state_machine/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type adminSubmitForm struct {
	UserId         int64  `json:"userId,omitempty"`
	Name           string `json:"name,omitempty"`
	Group          string `json:"group,omitempty"`
	AdditionalInfo string `json:"info,omitempty"`
}

const infoTemplate = "Имя: {{.Name}} \nГруппа: {{.Group}}\n{{if .AdditionalInfo}}Доп информация: {{.AdditionalInfo}} {{end}}"

type adminSubmitStartState struct {
	cache           interfaces.HandlersCache
	usersRepository interfaces.UsersRepository
	bot             *tgbotapi.BotAPI
}

func NewAdminSubmitState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, usersRepository interfaces.UsersRepository) *adminSubmitStartState {
	return &adminSubmitStartState{cache: cache, bot: bot, usersRepository: usersRepository}
}

func (*adminSubmitStartState) StateName() string {
	return constants.ADMIN_SUBMIT_START_STATE
}

func (state *adminSubmitStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	isAdmin, err := state.checkIfAdmin(message.From.ID)
	if err != nil {
		return err
	}
	if isAdmin {
		err = state.TransitionAndSend(interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE), tgbotapi.NewMessage(message.Chat.ID, "Вы уже админ группы"))
		return err
	}
	err = state.TransitionAndSend(interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_NAME_STATE), tgbotapi.NewMessage(message.Chat.ID, "Введите ваши фамилию и имя (Пример формата: Иванов Иван)"))
	return err
}

func (state *adminSubmitStartState) checkIfAdmin(tgId int64) (bool, error) {
	user, err := state.usersRepository.GetById(tgId)
	if err != nil {
		return false, err
	}
	return slices.Contains(user.Roles, entities.Admin), nil
}

func (state *adminSubmitStartState) TransitionAndSend(newState *interfaces.CachedInfo, msg tgbotapi.MessageConfig) error {
	err := state.cache.SaveState(*newState)
	if err != nil {
		return err
	}
	_, err = state.bot.Send(msg)
	return err
}

type adminSubmittingNameState struct {
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func NewAdminSubmittingNameState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmittingNameState {
	return &adminSubmittingNameState{cache: cache, bot: bot}
}

func (*adminSubmittingNameState) StateName() string {
	return constants.ADMIN_SUBMITTING_NAME_STATE
}

func (state *adminSubmittingNameState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Text == "" {
		return errors.New("no text in message")
	}
	info, err := json.Marshal(&adminSubmitForm{UserId: message.From.ID, Name: message.Text})
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(message.Chat.ID, string(info))
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_GROUP_STATE))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите ваш номер группы, указанный в ИИСе")
	_, err = state.bot.Send(msg)
	return err
}

type GroupsService interface {
	DoesGroupExist(string) (bool, error)
}

type adminSubmitingGroupState struct {
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
	srv   GroupsService
}

func NewAdminSubmitingGroupState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, srv GroupsService) *adminSubmitingGroupState {
	return &adminSubmitingGroupState{cache: cache, bot: bot, srv: srv}
}

func (*adminSubmitingGroupState) StateName() string {
	return constants.ADMIN_SUBMITTING_GROUP_STATE
}

func (state *adminSubmitingGroupState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Text == "" {
		return stateErrors.NewInvalidInput("No text in message")
	}
	exists, err := state.srv.DoesGroupExist(message.Text)
	if err != nil {
		return err
	}
	if !exists {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите номер существующей группы")
		_, err := state.bot.Send(msg)
		return err
	}
	info, err := state.cache.GetInfo(message.Chat.ID)
	if err != nil {
		return err
	}
	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}
	form.Group = message.Text

	marshalledInfo, err := json.Marshal(form)
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(message.Chat.ID, string(marshalledInfo))
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_PROOF_STATE))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Предоставьте доказательство вверенных группой полномочий (в виде фото, с дополнительной текстовой информацией по усмотрению)")
	_, err = state.bot.Send(msg)
	return err
}

type adminSubmittingProofState struct {
	cache    interfaces.HandlersCache
	bot      *tgbotapi.BotAPI
	requests interfaces.AdminRequestsRepository
}

func NewAdminSubmitingProofState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, requests interfaces.AdminRequestsRepository) *adminSubmittingProofState {
	return &adminSubmittingProofState{cache: cache, bot: bot, requests: requests}
}

func (state *adminSubmittingProofState) StateName() string {
	return constants.ADMIN_SUBMITTING_PROOF_STATE
}

func (state *adminSubmittingProofState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Photo == nil {
		_, err := state.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Отправьте фото как часть сообщения"))
		return err
	}
	info, err := state.cache.GetInfo(message.Chat.ID)
	if err != nil {
		return err
	}

	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return err
	}
	form.AdditionalInfo = message.Caption

	maxSizeId := selectMaxSizedPhoto(message.Photo)
	fileBytes, err := state.getFileBytes(maxSizeId)
	if err != nil {
		return err
	}

	err = state.cache.SaveState(*interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_WAITING_STATE))
	if err != nil {
		return err
	}

	msg := state.createTemplateResponse(message.Chat.ID, form, fileBytes)
	return state.sendPhotoToOwners(*msg, state.bot)
}

type adminWaitingState struct {
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func NewAdminWaitingProofState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminWaitingState {
	return &adminWaitingState{cache: cache, bot: bot}
}

func (state *adminWaitingState) StateName() string {
	return constants.ADMIN_WAITING_STATE
}

func (state *adminWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.From.ID, "Sorry, your last admin submit has not been proceeded yet")
	_, err := state.bot.Send(msg)
	return err
}

func createMarkupKeyboard(form *adminSubmitForm) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := constants.ADMIN_CALLBACKS + "accept" + fmt.Sprint(form.UserId)
	declineData := constants.ADMIN_CALLBACKS + "decline" + fmt.Sprint(form.UserId)
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Accept", acceptData), tgbotapi.NewInlineKeyboardButtonData("Decline", declineData))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}

func selectMaxSizedPhoto(sizes []tgbotapi.PhotoSize) string {
	maxSize := 0
	maxSizeId := ""
	for _, photo := range sizes {
		if photo.FileSize > maxSize {
			maxSizeId = photo.FileID
		}
	}
	return maxSizeId
}

func (state *adminSubmittingProofState) createTemplateResponse(chatId int64, form *adminSubmitForm, fileBytes []byte) *tgbotapi.PhotoConfig {
	msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileBytes{Name: "rnd_name", Bytes: fileBytes})
	var buf bytes.Buffer
	tmpl := template.Must(template.New("tmpl").Parse(infoTemplate))
	tmpl.Execute(&buf, form)
	msg.Caption = buf.String()
	msg.ReplyMarkup = createMarkupKeyboard(form)
	return &msg
}

func (state *adminSubmittingProofState) getFileBytes(fileId string) ([]byte, error) {
	file, err := state.bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		return nil, err
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

func (state *adminSubmittingProofState) sendPhotoToOwners(msg tgbotapi.PhotoConfig, bot *tgbotapi.BotAPI) error {
	owners := strings.Split(os.Getenv("OWNERS"), ",")

	for _, owner := range owners {
		chatId, err := strconv.ParseInt(owner, 10, 64)
		if err != nil {
			return errors.Join(err, fmt.Errorf("invalid owner id value %s", owner))
		}
		msg.ChatID = chatId
		sentMsg, err := bot.Send(msg)
		if err != nil {
			return err
		}
		err = state.requests.SaveRequest(interfaces.NewAdminRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, uuid.NewString()))
		if err != nil {
			return err
		}
	}
	return nil
}
