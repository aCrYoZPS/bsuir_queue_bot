package stateMachine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var errInvalidInput error
var errWrongGroup error

const (
	ADMIN_SUBMIT_START_STATE     StateName = "submit"
	ADMIN_SUBMITTING_NAME_STATE  StateName = "submitting_name"
	ADMIN_SUBMITTING_GROUP_STATE StateName = "submitting_group"
	ADMIN_SUBMITTING_PROOF_STATE StateName = "submitting_proof"
	ADMIN_WAITING_STATE          StateName = "waiting"
)

type adminSubmitForm struct {
	ChatId         int64  `json:"chatId,omitempty"`
	Name           string `json:"name,omitempty"`
	Group          string `json:"group,omitempty"`
	AdditionalInfo string `json:"info,omitempty"`
}

const infoTemplate = "Имя: {{.Name}} \nГруппа: {{.Group}}\n{{if .AdditionalInfo}}Доп информация: {{.AdditionalInfo}} {{end}}"

type adminSubmitStartState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func newAdminSubmitState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmitStartState {
	return &adminSubmitStartState{cache: cache, bot: bot}
}

func (*adminSubmitStartState) StateName() StateName {
	return ADMIN_SUBMIT_START_STATE
}

func (state *adminSubmitStartState) Handle(chatId int64, message *tgbotapi.Message) error {
	err := state.cache.SaveState(*interfaces.NewCachedInfo(chatId, string(ADMIN_SUBMITTING_NAME_STATE)))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "Введите ваши фамилию и имя (Пример формата: Иван Иванов)")
	_, err = state.bot.Send(msg)
	return err
}

type adminSubmittingNameState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func newAdminSubmittingNameState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmittingNameState {
	return &adminSubmittingNameState{cache: cache, bot: bot}
}

func (*adminSubmittingNameState) StateName() StateName {
	return ADMIN_SUBMITTING_NAME_STATE
}

func (state *adminSubmittingNameState) Handle(chatId int64, message *tgbotapi.Message) error {
	if message.Text == "" {
		return errors.New("no text in message")
	}
	info, err := json.Marshal(&adminSubmitForm{ChatId: message.Chat.ID, Name: message.Text})
	if err != nil {
		return err
	}
	err = state.cache.SaveInfo(chatId, string(info))
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, string(ADMIN_SUBMITTING_GROUP_STATE)))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "Введите ваш номер группы, указанный в ИИСе")
	_, err = state.bot.Send(msg)
	return err
}

type adminSubmitingGroupState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
	srv   GroupsService
}

func newAdminSubmitingGroupState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI, srv GroupsService) *adminSubmitingGroupState {
	return &adminSubmitingGroupState{cache: cache, bot: bot, srv: srv}
}

func (*adminSubmitingGroupState) StateName() StateName {
	return ADMIN_SUBMITTING_GROUP_STATE
}

func (state *adminSubmitingGroupState) Handle(chatId int64, message *tgbotapi.Message) error {
	if message.Text == "" {
		return errors.Join(errInvalidInput, errors.New("no text in message"))
	}
	exists, err := state.srv.DoesGroupExist(message.Text)
	if err != nil {
		return err
	}
	if !exists {
		msg := tgbotapi.NewMessage(chatId, "Введите номер существующей группы")
		_, err := state.bot.Send(msg)
		return err
	}
	info, err := state.cache.GetInfo(chatId)
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
	err = state.cache.SaveInfo(chatId, string(marshalledInfo))
	if err != nil {
		return err
	}
	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, string(ADMIN_SUBMITTING_PROOF_STATE)))
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "Предоставьте доказательство вверенных группой полномочий (в виде фото, с дополнительной текстовой информацией по усмотрению)")
	_, err = state.bot.Send(msg)
	return err
}

type adminSubmittingProofState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func newAdminSubmitingProofState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmittingProofState {
	return &adminSubmittingProofState{cache: cache, bot: bot}
}

func (state *adminSubmittingProofState) StateName() StateName {
	return ADMIN_SUBMITTING_PROOF_STATE
}

func (state *adminSubmittingProofState) Handle(chatId int64, message *tgbotapi.Message) error {
	if message.Photo == nil {
		return errors.Join(errInvalidInput, errors.New("no photo"))
	}
	info, err := state.cache.GetInfo(chatId)
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
	file, err := state.bot.GetFile(tgbotapi.FileConfig{FileID: maxSizeId})
	if err != nil {
		return err
	}
	link := file.Link(os.Getenv("BOT_TOKEN"))
	resp, err := http.Get(link)
	if err != nil && err != io.EOF {
		return err
	}
	defer resp.Body.Close()

	err = state.cache.SaveState(*interfaces.NewCachedInfo(chatId, string(ADMIN_WAITING_STATE)))
	if err != nil {
		return err
	}

	msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileReader{Name: "rnd_name", Reader: resp.Body})
	var buf bytes.Buffer
	tmpl := template.Must(template.New("tmpl").Parse(infoTemplate))
	tmpl.Execute(&buf, form)
	msg.Caption = buf.String()
	msg.ReplyMarkup = createMarkupKeyboard(form)
	return tgutils.SendPhotoToOwners(msg, state.bot)
}

type adminWaitingState struct {
	State
	cache interfaces.HandlersCache
	bot   *tgbotapi.BotAPI
}

func newAdminWaitingProofState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminWaitingState {
	return &adminWaitingState{cache: cache, bot: bot}
}

func (state *adminWaitingState) StateName() StateName {
	return ADMIN_WAITING_STATE
}

func (state *adminWaitingState) Handle(chatId int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(chatId, "Sorry, your last admin submit has not been proceeded yet")
	_, err := state.bot.Send(msg)
	return err
}

func createMarkupKeyboard(form *adminSubmitForm) *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	acceptData := ADMIN_CALLBACKS + "accept" + fmt.Sprint(form.ChatId)
	declineData := ADMIN_CALLBACKS + "decline" + fmt.Sprint(form.ChatId)
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
