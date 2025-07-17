package stateMachine

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"text/template"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var errInvalidInput error

const (
	ADMIN_SUBMIT_START_STATE     StateName = "submit"
	ADMIN_SUBMITTING_NAME_STATE  StateName = "submitting_name"
	ADMIN_SUBMITTING_GROUP_STATE StateName = "submitting_group"
	ADMIN_SUBMITTING_PROOF_STATE StateName = "submitting_proof"
)

type adminSubmitForm struct {
	Name           string `json:"name,omitempty"`
	Group          string `json:"group,omitempty"`
	AdditionalInfo string `json:"info,omitempty"`
}

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
	info, err := json.Marshal(&adminSubmitForm{Name: message.Text})
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
}

func newAdminSubmitingGroupState(cache interfaces.HandlersCache, bot *tgbotapi.BotAPI) *adminSubmitingGroupState {
	return &adminSubmitingGroupState{cache: cache, bot: bot}
}

func (*adminSubmitingGroupState) StateName() StateName {
	return ADMIN_SUBMITTING_GROUP_STATE
}

func (state *adminSubmitingGroupState) Handle(chatId int64, message *tgbotapi.Message) error {
	if message.Text == "" {
		return errors.Join(errInvalidInput, errors.New("no text in message"))
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
	var buf bytes.Buffer
	tmplString := "Имя: {{.Name}} \nГруппа: {{.Group}}\n{{if .AdditionalInfo}}Доп информация: {{.AdditionalInfo}} {{end}}"
	tmpl := template.Must(template.New("tmpl").Parse(tmplString))
	tmpl.Execute(&buf, form)

	maxSize := 0
	maxSizeId := ""
	for _, photo := range message.Photo {
		if photo.FileSize > maxSize {
			maxSizeId = photo.FileID
		}
	}
	file, err := state.bot.GetFile(tgbotapi.FileConfig{FileID: maxSizeId})
	if err != nil {
		return err
	}

	link := file.Link(os.Getenv("BOT_TOKEN"))
	filePath := os.Getenv("DATA_DIRECTORY") + file.FileUniqueID
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(filePath)

	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileReader{Name: "rnd_name", Reader: resp.Body})
	msg.Caption = buf.String()
	msg.ReplyMarkup = createMarkupKeyboard()
	err = tgutils.SendPhotoToOwners(msg, state.bot)
	return err
}

func createMarkupKeyboard() *tgbotapi.InlineKeyboardMarkup {
	row := []tgbotapi.InlineKeyboardButton{}
	row = append(row, tgbotapi.NewInlineKeyboardButtonData("Accept", ADMIN_CALLBACKS+"accept"), tgbotapi.NewInlineKeyboardButtonData("Decline", ADMIN_CALLBACKS+"decline"))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return &keyboard
}
