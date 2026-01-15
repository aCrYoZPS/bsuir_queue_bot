package delete

import (
	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
	tgutils "github.com/aCrYoZPS/bsuir_queue_bot/src/utils/tg_utils"
)

type QueueService interface {

}

type DeleteStartState struct {
	cache interfaces.HandlersCache
	bot   *tgutils.Bot
}

func NewDeleteStartState(cache interfaces.HandlersCache, bot *tgutils.Bot) *DeleteStartState {
	return &DeleteStartState{cache: cache, bot: bot}
}

func (state *DeleteStartState) StateName() string {
	return constants.DELETE_START_STATE
}

func (state *DeleteStartState) Handle() {
	
}
