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
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type adminSubmitForm struct {
	UserId         int64  `json:"userId,omitempty"`
	Name           string `json:"name,omitempty"`
	Group          string `json:"group,omitempty"`
	TgName         string `json:"tg_name,omitempty"`
	AdditionalInfo string `json:"info,omitempty"`
}

const infoTemplate = "(ЗАЯВКА НА РОЛЬ АДМИНИСТРАТОРА)\n Имя: {{.Name}} \nГруппа: {{.Group}}\nИмя пользователя: @{{.TgName}} \n{{if .AdditionalInfo}}Доп информация: {{.AdditionalInfo}} {{end}}"

type adminSubmitStartState struct {
	cache           interfaces.HandlersCache
	usersRepository interfaces.UsersRepository
	bot             *tgutils.Bot
}

func NewAdminSubmitState(cache interfaces.HandlersCache, bot *tgutils.Bot, usersRepository interfaces.UsersRepository) *adminSubmitStartState {
	return &adminSubmitStartState{cache: cache, bot: bot, usersRepository: usersRepository}
}

func (*adminSubmitStartState) StateName() string {
	return constants.ADMIN_SUBMIT_START_STATE
}

func (state *adminSubmitStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.usersRepository.GetByTgId(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("couldn't get user by id when checking admin: %w", err)
	}
	if slices.Contains(user.Roles, entities.Admin) {
		err = state.TransitionAndSend(ctx, interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE), tgbotapi.NewMessage(message.Chat.ID, "Вы уже админ группы"))
		return err
	}
	if user.Id != 0 {
		info, err := json.Marshal(&adminSubmitForm{UserId: message.From.ID, Group: user.GroupName, Name: user.FullName, TgName: message.From.UserName})
		if err != nil {
			return fmt.Errorf("failed to marshal admin submit form: %w", err)
		}
		err = state.cache.SaveInfo(ctx, message.Chat.ID, string(info))
		if err != nil {
			return fmt.Errorf("failed to save info to cache during submitting name for admin: %w", err)
		}
		err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_PROOF_STATE))
		if err != nil {
			return fmt.Errorf("failed to transition to submitting proof state during admin submit start state handling: %w", err)
		}
		return nil
	}
	err = state.TransitionAndSend(ctx, interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_NAME_STATE),
		tgbotapi.NewMessage(message.Chat.ID, "Введите ваши фамилию и имя (Пример формата: Иванов Иван)"))
	return err
}

func (state *adminSubmitStartState) TransitionAndSend(ctx context.Context, newState *interfaces.CachedInfo, msg tgbotapi.MessageConfig) error {
	err := state.cache.SaveState(ctx, *newState)
	if err != nil {
		return fmt.Errorf("couldn't save state during admin submit: %w", err)
	}
	_, err = state.bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("couldn't send message during admin submit: %w", err)
	}
	return nil
}

func (state *adminSubmitStartState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.RemoveInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove info during admin submit start state reversal: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to idle state during admin submit start state reversal: %w", err)
	}
	return nil
}

type StateMachine interface {
	HandleState(ctx context.Context, msg *tgbotapi.Message) error
}

type adminSubmittingNameState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	machine StateMachine
}

func NewAdminSubmittingNameState(cache interfaces.HandlersCache, bot *tgutils.Bot, machine StateMachine) *adminSubmittingNameState {
	return &adminSubmittingNameState{cache: cache, bot: bot, machine: machine}
}

func (*adminSubmittingNameState) StateName() string {
	return constants.ADMIN_SUBMITTING_NAME_STATE
}

func (state *adminSubmittingNameState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Text == "" {
		return errors.New("no text in message")
	}

	info, err := json.Marshal(&adminSubmitForm{UserId: message.From.ID, Name: message.Text, TgName: message.From.UserName})
	if err != nil {
		return fmt.Errorf("failed to marshal admin submit form: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(info))
	if err != nil {
		return fmt.Errorf("failed to save info to cache during submitting name for admin: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_GROUP_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during transition from admin submitting name to submitting group: %w", err)
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите ваш номер группы, указанный в ИИСе")
	_, err = state.bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message during admin submitting name: %w", err)
	}
	return nil
}

func (state *adminSubmittingNameState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.ADMIN_SUBMIT_START_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to admin submit start state during admin submitting name state reversal: %w", err)
	}
	msg.Text = "/assign"
	err = state.machine.HandleState(ctx, msg)
	return err
}

type GroupsService interface {
	DoesGroupExist(ctx context.Context, groupname string) (bool, error)
}

type adminSubmitingGroupState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	srv     GroupsService
	machine StateMachine
}

func NewAdminSubmitingGroupState(cache interfaces.HandlersCache, bot *tgutils.Bot, srv GroupsService, machine StateMachine) *adminSubmitingGroupState {
	return &adminSubmitingGroupState{cache: cache, bot: bot, srv: srv, machine: machine}
}

func (*adminSubmitingGroupState) StateName() string {
	return constants.ADMIN_SUBMITTING_GROUP_STATE
}

func (state *adminSubmitingGroupState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Text == "" {
		return stateErrors.NewInvalidInput("no text in message")
	}

	exists, err := state.srv.DoesGroupExist(ctx, message.Text)
	if err != nil {
		return fmt.Errorf("failed to check if group exists during admin submitting group state: %w", err)
	}
	if !exists {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите номер существующей в ИИСе группы")
		_, err := state.bot.SendCtx(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to send group not exists message during admin submitting group: %w", err)
		}
		return nil
	}

	info, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info from cache during admin submitting group")
	}
	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info from cache during admin submitting group: %w", err)
	}
	form.Group = message.Text
	marshalledInfo, err := json.Marshal(form)
	if err != nil {
		return fmt.Errorf("failed marshal info during admin submitting group: %w", err)
	}

	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(marshalledInfo))
	if err != nil {
		return fmt.Errorf("failed to save info to cache during admin submitting group")
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_SUBMITTING_PROOF_STATE))
	if err != nil {
		return fmt.Errorf("failed save new state during transitioning from admin submitting group to admin submitting proof")
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Предоставьте доказательство вверенных группой полномочий (в виде фото, с дополнительной текстовой информацией по усмотрению)")
	_, err = state.bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message at the end of admin submitting group")
	}
	return nil
}

func (state *adminSubmitingGroupState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.ADMIN_SUBMITTING_NAME_STATE))
	if err != nil {
		return fmt.Errorf("failed to save admin submitting name state during admin submitting group state reversal: %w", err)
	}
	msg.Text = ""
	err = state.machine.HandleState(ctx, msg)
	return err
}

type adminSubmittingProofState struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	requests interfaces.AdminRequestsRepository
	machine  StateMachine
}

func NewAdminSubmitingProofState(cache interfaces.HandlersCache, bot *tgutils.Bot, requests interfaces.AdminRequestsRepository, machine StateMachine) *adminSubmittingProofState {
	return &adminSubmittingProofState{cache: cache, bot: bot, requests: requests, machine: machine}
}

func (state *adminSubmittingProofState) StateName() string {
	return constants.ADMIN_SUBMITTING_PROOF_STATE
}

func (state *adminSubmittingProofState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	if message.Photo == nil {
		_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Отправьте фото как часть сообщения"))
		if err != nil {
			return fmt.Errorf("failed to send no photo message during admin submitting proof: %w", err)
		}
		return nil
	}
	info, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info during admin submitting proof: %w", err)
	}

	form := &adminSubmitForm{}
	err = json.Unmarshal([]byte(info), form)
	if err != nil {
		return fmt.Errorf("failed to unmarshal form during admin submitting proof: %w", err)
	}
	form.AdditionalInfo = message.Caption

	maxSizeId := selectMaxSizedPhoto(message.Photo)
	fileBytes, err := state.getFileBytes(maxSizeId)
	if err != nil {
		return fmt.Errorf("failed to get file bytes for photo during admin submitting proof: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.ADMIN_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during transitioning from admin proof to admin waiting")
	}

	msg := state.createTemplateResponse(message.Chat.ID, form, fileBytes)
	return state.sendPhotoToOwners(ctx, message.Chat.ID, *msg, state.bot)
}

func (state *adminSubmittingProofState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(msg.Chat.ID, constants.ADMIN_SUBMITTING_GROUP_STATE))
	if err != nil {
		return fmt.Errorf("failed to transition to admin submitting group state during reversal of admin submitting proof state: %w", err)
	}
	msg.Text = ""
	err = state.machine.HandleState(ctx, msg)
	return err
}

type adminWaitingState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
}

func NewAdminWaitingProofState(cache interfaces.HandlersCache, bot *tgutils.Bot) *adminWaitingState {
	return &adminWaitingState{cache: cache, bot: bot}
}

func (state *adminWaitingState) StateName() string {
	return constants.ADMIN_WAITING_STATE
}

func (state *adminWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.From.ID, "Подождите,ваш запрос на роль администратора ещё обрабатывается")
	_, err := state.bot.SendCtx(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message to user during admin waiting state: %w", err)
	}
	return nil
}

func (state *adminWaitingState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	return nil
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

func (state *adminSubmittingProofState) sendPhotoToOwners(ctx context.Context, senderChatId int64, msg tgbotapi.PhotoConfig, bot *tgutils.Bot) error {
	owners := strings.Split(os.Getenv("OWNERS"), ",")

	for _, owner := range owners {
		chatId, err := strconv.ParseInt(owner, 10, 64)
		if err != nil {
			return errors.Join(err, fmt.Errorf("invalid owner id value %s", owner))
		}
		msg.ChatID = chatId
		sentMsg, err := bot.SendCtx(ctx, msg)
		if err != nil {
			if errors.Is(err, tgutils.ErrMsgInvalidLen) {
				_, err := bot.SendCtx(ctx, tgbotapi.NewMessage(senderChatId, "Ваша заявка превысила допустимый лимит длины. Пожалуйста,перепишите её и отправьте снова"))
				if err != nil {
					return fmt.Errorf("failed to send too long response during admin submitting proof state: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to send msg to owner id %s during admin proof submit: %w", owner, err)
		}
		err = state.requests.SaveRequest(ctx, interfaces.NewAdminRequest(int64(sentMsg.MessageID), sentMsg.Chat.ID, uuid.NewString()))
		if err != nil {
			return err
		}
	}
	return nil
}
