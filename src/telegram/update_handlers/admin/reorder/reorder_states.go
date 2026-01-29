package reorder

import (
	"context"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/entities"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LessonsRepository interface {
	GetSubjects(ctx context.Context, groupId int64) ([]string, error)
}

type UsersRepository interface {
	GetByTgId(ctx context.Context, tgId int64) (entities.User, error)
}

type ReorderStartState struct {
	cache   interfaces.HandlersCache
	bot     *tgutils.Bot
	lessons LessonsRepository
	users   UsersRepository
}

func NewReorderStartState(cache interfaces.HandlersCache, bot *tgutils.Bot, users UsersRepository, lessons LessonsRepository) *ReorderStartState {
	return &ReorderStartState{cache: cache, bot: bot, users: users, lessons: lessons}
}

func (state *ReorderStartState) Handle(ctx context.Context, message *tgbotapi.Message) error {
	return nil
}
