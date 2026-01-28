package delete

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StudentsRepository interface {
	GetStudents(ctx context.Context, groupname string) ([]entities.User, error)
	GetByTgId(ctx context.Context, tgId int64) (*entities.User, error)
	Delete(ctx context.Context, id int64) error
}

type DeleteStartState struct {
	bot      *tgutils.Bot
	students StudentsRepository
	cache    interfaces.HandlersCache
}

type StudentId = int64

type DeleteInfo struct {
	Students map[int]StudentId
}

func NewDeleteStartState(bot *tgutils.Bot, students StudentsRepository, cache interfaces.HandlersCache) *DeleteStartState {
	return &DeleteStartState{bot: bot, students: students, cache: cache}
}

func (state *DeleteStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	usr, err := state.students.GetByTgId(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user by tg id in delete start state: %w", err)
	}
	students, err := state.students.GetStudents(ctx, usr.GroupName)
	if err != nil {
		return fmt.Errorf("failed to get students in delete start state: %w", err)
	}

	info := &DeleteInfo{Students: map[int]StudentId{}}
	for i, student := range students {
		info.Students[i+1] = student.Id
	}
	jsonedInfo, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal state info in delete start state: %w", err)
	}
	err = state.cache.SaveInfo(ctx, message.Chat.ID, string(jsonedInfo))
	if err != nil {
		return fmt.Errorf("failed to save delete start state info: %w", err)
	}

	resp := tgbotapi.NewMessage(message.Chat.ID, "Выберите студента для удаления из списка:\n"+state.formatStudentsOutput(students))
	_, err = state.bot.SendCtx(ctx, resp)
	if err != nil {
		return fmt.Errorf("failed to send response during delete start state: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.DELETE_CHOOSE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save state during delete start state: %w", err)
	}
	return nil
}

func (state *DeleteStartState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}

func (state *DeleteStartState) formatStudentsOutput(students []entities.User) string {
	builder := strings.Builder{}
	for i, student := range students {
		fmt.Fprintf(&builder, "%d. %s\n", i+1, student.FullName)
	}
	return builder.String()
}

type DeleteChooseState struct {
	bot      *tgutils.Bot
	cache    interfaces.HandlersCache
	students StudentsRepository
}

func NewDeleteChooseState(bot *tgutils.Bot, cache interfaces.HandlersCache, students StudentsRepository) *DeleteChooseState {
	return &DeleteChooseState{bot: bot, cache: cache, students: students}
}

func (state *DeleteChooseState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	num, err := strconv.Atoi(message.Text)
	if err != nil {
		_, sendErr := state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите конкретную цифру с участником группы"))
		if sendErr != nil {
			return fmt.Errorf("failed to send response in delete choose state: %w", sendErr)
		}
		return fmt.Errorf("failed to parse number from user input in delete choose state: %w", err)
	}

	jsonedInfo, err := state.cache.GetInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get info from cache in delete choose state: %w", err)
	}
	var info DeleteInfo
	err = json.NewDecoder(strings.NewReader(jsonedInfo)).Decode(&info)
	if err != nil {
		return fmt.Errorf("failed to decode delete info ")
	}
	student, exists := info.Students[num]
	if !exists {
		_, err = state.bot.SendCtx(ctx, tgbotapi.NewMessage(message.Chat.ID, "Введите корректный номер с участником группы"))
		if err != nil {
			return fmt.Errorf("failed to send response in delete choose state: %w", err)
		}
		return fmt.Errorf("failed to parse number from user input in delete choose state: %w", err)
	}

	err = state.students.Delete(ctx, student)
	if err != nil {
		return fmt.Errorf("failed to delete user %d: %w", student, err)
	}

	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to save idle state in delete choose state: %w", err)
	}
	return nil
}

func (state *DeleteChooseState) Revert(ctx context.Context, message *tgbotapi.Message) error {
	err := state.cache.RemoveInfo(ctx, message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove info during delete choose state reversal: %w", err)
	}
	err = state.cache.SaveState(ctx, *interfaces.NewCachedInfo(message.Chat.ID, constants.IDLE_STATE))
	if err != nil {
		return fmt.Errorf("failed to change state during delete choose state reversal: %w", err)
	}
	return nil
}
