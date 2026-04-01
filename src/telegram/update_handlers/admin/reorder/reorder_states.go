package reorder

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	iis_api_entities "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/sqlite/persistance"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LessonsRepository interface {
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
	GetNext(ctx context.Context, subject string, groupId int64) ([]persistance.Lesson, error)
	Get(ctx context.Context, id int64) (persistance.Lesson, error)
}

type UsersRepository interface {
	GetByTgId(ctx context.Context, tgId int64) (entities.User, error)
}

type GroupsRepository interface {
	GetById(ctx context.Context, id int) (*iis_api_entities.Group, error)
}

type ReorderStartState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	lessons LessonsRepository
	users   UsersRepository
	groups  GroupsRepository
}

func NewReorderStartState(cache interfaces.HandlersCache, bot *tgutils.Bot, users UsersRepository, lessons LessonsRepository, groups GroupsRepository) *ReorderStartState {
	return &ReorderStartState{cache: cache, bot: bot, users: users, lessons: lessons, groups: groups}
}

type ReorderInfo struct {
	MarkupMessageId int    `json:"markup_id,omitempty"`
	Subject         string `json:"subject,omitempty"`
	AllLessons      bool   `json:"all,omitempty"`
	//Should equal 0, if all lessons flag is set
	LessonId  int64  `json:"lesson_id,omitempty"`
	GroupName string `json:"groupname,omitempty"`
}

func (state *ReorderStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	user, err := state.users.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user during reorder start state: %w", err)
	}

	userGroup, err := state.groups.GetById(ctx, int(user.GroupId))
	if err != nil {
		return fmt.Errorf("failed to group id during reorder state: %w", err)
	}

	subjects, err := state.lessons.GetSubjects(ctx, user.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get subjects during reorder start state: %w", err)
	}
	markup := state.markupFromSubjects(subjects)
	resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите предмет для изменения порядка очереди")
	resp.ReplyMarkup = markup
	sended, err := state.bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send response during reorder start state: %w", err)
	}

	jsonedInfo, err := json.Marshal(&ReorderInfo{MarkupMessageId: sended.MessageID, GroupName: userGroup.Name})
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedInfo))
	if err != nil {
		return fmt.Errorf("failed to save info during request reorder start state %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_WAITING_STATE))
	if err != nil {
		return fmt.Errorf("faield to save next state during request reorder start state: %w", err)
	}
	return nil
}

func (state *ReorderStartState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.RemoveInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove info during reorder start state: %w", err)
	}
	return nil
}

func (state *ReorderStartState) markupFromSubjects(subjects []string) tgbotapi.InlineKeyboardMarkup {
	var markup tgbotapi.InlineKeyboardMarkup
	for chunk := range slices.Chunk(subjects, 3) {
		var row []tgbotapi.InlineKeyboardButton
		for _, subject := range chunk {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(subject, lessonCallback(subject)))
		}
		markup.InlineKeyboard = append(markup.InlineKeyboard, row)
	}
	return markup
}

func lessonCallback(subject string) (callback string) {
	return constants.REORDER_LESSON_NAME_CALLBACK + "|" + subject
}

func parseLessonCallback(callback string) (subject string) {
	return strings.TrimPrefix(callback, constants.REORDER_LESSON_NAME_CALLBACK+"|")
}

type StateMachine interface {
	Handle(ctx context.Context, message *tgbotapi.Message) error
}

type ReorderWaitingState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	machine StateMachine
}

func NewReorderWaitingState(cache interfaces.HandlersCache, bot *tgutils.Bot) *ReorderWaitingState {
	return &ReorderWaitingState{cache: cache, bot: bot}
}

func (state *ReorderWaitingState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, выберите предмет для изменения порядка очередности"))
	if err != nil {
		return fmt.Errorf("failed to send response during reorder waiting state: %w", err)
	}
	return nil
}

func (state *ReorderWaitingState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info from database during reorder waiting state reversal: %w", err)
	}
	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info: %w", err)
	}

	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, info.MarkupMessageId,
		tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to edit markup in reorder waiting state reversal: %w", err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_REQUEST_START_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during reorder waiting state reeversal: %w", err)
	}

	err = state.machine.Handle(ctx, message)
	return err
}

type ReorderChooseAllState struct {
	bot     *tgutils.Bot
	cache   interfaces.HandlersCache
	machine StateMachine
	lessons LessonsRepository
	users   UsersRepository
}

func NewReorderChooseState(bot *tgutils.Bot, cache interfaces.HandlersCache, machine StateMachine, lessons LessonsRepository, users UsersRepository) *ReorderChooseAllState {
	return &ReorderChooseAllState{bot: bot, cache: cache, lessons: lessons, machine: machine, users: users}
}

func orderationMessage(chatId int64) tgbotapi.MessageConfig {
	text := "Выберите способы сортировки данных,через запятую (порядок важен, сортировка будет применена в указанном порядке).\n1 - по времени отправки. 2 - по номеру лабораторной. Добавьте префикс + к номеру, если хотите установить сортировку по убыванию"
	return tgbotapi.NewMessage(chatId, text)
}

// WRITE CHOOSE LESSON STATE, FOR CASES WHEN WE WANT TO ORDERATE A SPECIFIC ONE
func (state *ReorderChooseAllState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_REQUEST_METHOD_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during reorder choose state handling: %w", err)
	}
	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info during reorder choose state: %w", err)
	}
	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json into reorder info: %w", err)
	}
	if message.Text == "Да" {
		info.AllLessons = true
		jsonInfo, err := json.Marshal(info)
		if err != nil {
			return fmt.Errorf("failed to marshal info into json in reorder choose state: %w", err)
		}
		err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonInfo))
		if err != nil {
			return fmt.Errorf("failed to sace info during reorder choose state: %w", err)
		}
	} else {
		usr, err := state.users.GetByTgId(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("failed to get user by tg id during reorder choose all state: %w", err)
		}
		next, err := state.lessons.GetNext(ctx, info.Subject, usr.GroupId)
		if err != nil {
			return fmt.Errorf("failed to get next lessons during reorder choose all state: %w", err)
		}
		var keyboard tgbotapi.InlineKeyboardMarkup

		for chunk := range slices.Chunk(next, 3) {
			var row []tgbotapi.InlineKeyboardButton
			for _, lesson := range chunk {
				buttonVisual := fmt.Sprintf("%s %s", lesson.Subject, lesson.DateTime.Format("02.01.2006"))
				if lesson.SubgroupNumber != 0 {
					buttonVisual += fmt.Sprintf(" (%d)", lesson.SubgroupNumber)
				}
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(buttonVisual, constants.REORDER_LESSON_CONCRETE_CALLBACK))
			}
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}

		err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_CHOOSE_LESSON_STATE))
		if err != nil {
			return fmt.Errorf("failed to save choose lesson state during reorder choose state: %w", err)
		}

		resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите занятие для изменения сортировки")
		resp.ReplyMarkup = keyboard
		sentResponse, err := state.bot.SendCtx(ctx, resp)
		if err != nil {
			return fmt.Errorf("failed to send response during reorder choose all state: %w", err)
		}

		info.MarkupMessageId = sentResponse.MessageID
		jsonedInfo, err := json.Marshal(&info)
		if err != nil {
			return fmt.Errorf("failed to convert reorder info into json during reorder choose all state: %w", err)
		}
		err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedInfo))
		if err != nil {
			return fmt.Errorf("failed to save info during requesr reorder choose all state: %w", err)
		}
	}
	_, err = state.bot.SendCtx(ctx, orderationMessage(message.Chat.ID))
	if err != nil {
		return fmt.Errorf("faield to send response during reorder choose state: %w", err)
	}
	return nil
}

func createLessonConcreteCallback(lesson persistance.Lesson) string {
	return fmt.Sprintf("%s|%d", constants.REORDER_LESSON_CONCRETE_CALLBACK, lesson.Id)
}

func parseLessonConcreteCallback(callback string) (id int64, err error) {
	return strconv.ParseInt(callback, 10, 64)
}

func (state *ReorderChooseAllState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_REQUEST_START_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during reorder choose state: %w", err)
	}
	err = state.machine.Handle(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

type ReorderChooseLessonState struct {
	cache interfaces.HandlersCache
	bot   *tgutils.Bot
}

func NewReorderChooseLessonState(cache interfaces.HandlersCache, bot *tgutils.Bot) *ReorderChooseLessonState {
	return &ReorderChooseLessonState{
		cache: cache,
		bot:   bot,
	}
}

func (state *ReorderChooseLessonState) Handle(ctx context.Context, msg *tgbotapi.Message) error {
	_, err := state.bot.SendCtx(ctx, tgbotapi.NewMessage(msg.Chat.ID, "Пожалуйста, выберите занятие для изменения сортировки"))
	if err != nil {
		return fmt.Errorf("failed to send response during reorder choose lesson state: %w", err)
	}
	return nil
}

func (state *ReorderChooseLessonState) Revert(ctx context.Context, msg *tgbotapi.Message) error {
	jsonedInfo, err := state.cache.GetInfo(ctx, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info during reorder choose lesson state reversal: %w", err)
	}

	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal jsoned info into reorder states info: %w", err)
	}

	_, err = state.bot.SendCtx(ctx, tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, info.MarkupMessageId, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{})))
	if err != nil {
		return fmt.Errorf("failed to remove reply markup during reorder choose lesson state: %w", err)
	}
	return nil
}

type SheetsApi interface {
	ReorderLessons(ctx context.Context, orderTypes []entities.OrderType, groupName, subject string) error
	ReorderLesson(ctx context.Context, orderTypes []entities.OrderType, groupName string, lesson persistance.Lesson) error
}

type ReorderMethodState struct {
	cache    interfaces.HandlersCache
	bot      *tgutils.Bot
	requests LessonRequestsRepository
	sheets   SheetsApi
	lessons  LessonsRepository
	machine  StateMachine
}

type LessonRequestsRepository interface {
	ChangeOrderation(context.Context, []entities.OrderType, int64) error
	ChangeSubjectOrderation(context.Context, []entities.OrderType, string) error
}

func NewReorderMethodState(cache interfaces.HandlersCache, bot *tgutils.Bot, requests LessonRequestsRepository, sheets SheetsApi) *ReorderMethodState {
	return &ReorderMethodState{cache: cache, bot: bot, requests: requests, sheets: sheets}
}

func (state *ReorderMethodState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	methods, err := state.parseMessage(message)
	if err != nil {
		return fmt.Errorf("failed to parse reorder method message: %w", err)
	}
	orderTypes := []entities.OrderType{}
	for _, method := range methods {
		orderTypes = append(orderTypes, entities.OrderType{Ascending: method.ascending, Value: entities.OrderField(method.orderation)})
	}
	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get jsoned info in reorder method state: %w", err)
	}
	var info ReorderInfo
	err = json.Unmarshal([]byte(jsonedInfo), &info)
	if err != nil {
		return fmt.Errorf("failed to unmarshal info during reorder method state: %w", err)
	}

	if info.AllLessons {
		err = state.requests.ChangeSubjectOrderation(ctx, orderTypes, info.Subject)
		if err != nil {
			return fmt.Errorf("failed to change orderation of subject in db during reorder method state: %w", err)
		}
		err = state.sheets.ReorderLessons(ctx, orderTypes, info.GroupName, info.Subject)
		if err != nil {
			return fmt.Errorf("failed to reorder subject in google sheets during reorder method state: %w", err)
		}
	} else {
		err = state.requests.ChangeOrderation(ctx, orderTypes, info.LessonId)
		if err != nil {
			return fmt.Errorf("failed to change orderation of lesson in db during reorder method state: %w", err)
		}

		lesson, err := state.lessons.Get(ctx, info.LessonId)
		if err != nil {
			return fmt.Errorf("failed to get lesson by id during reorder method state: %w", err)
		}

		err = state.sheets.ReorderLesson(ctx,orderTypes, info.GroupName, lesson)
		if err != nil {
			return fmt.Errorf("failed to reorder lesson in google sheets during reorder method state: %w", err)
		}
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state during reorder method state: %w", err)
	}

	return nil
}

func (state *ReorderMethodState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.REORDER_CHOOSE_STATE))
	if err != nil {
		return err
	}
	return state.machine.Handle(ctx, message)
}

func (state *ReorderMethodState) parseMessage(message *tgbotapi.Message) ([]struct {
	orderation int8
	ascending  bool
}, error) {
	returned := []struct {
		orderation int8
		ascending  bool
	}{}
	parts := strings.Split(message.Text, ",")
	for _, part := range parts {
		after, found := strings.CutPrefix(part, "+")
		order, err := strconv.ParseInt(after, 10, 8)
		if err != nil {
			return nil, err
		}
		returned = append(returned, struct {
			orderation int8
			ascending  bool
		}{orderation: int8(order), ascending: !found})
	}
	return returned, nil
}
