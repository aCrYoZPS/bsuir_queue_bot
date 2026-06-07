package interfaces

import (
	"context"
	"sync"
	"time"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers/constants"
)

type CachedInfo struct {
	chatId   int64
	state    constants.State
	sendTime time.Time
}

func (info *CachedInfo) State() string {
	return string(info.state)
}

func (info *CachedInfo) SendTime() time.Time {
	return info.sendTime
}

func (info *CachedInfo) ChatId() int64 {
	return info.chatId
}

func NewCachedInfo(ChatId int64, state constants.State) *CachedInfo {
	return &CachedInfo{
		chatId:   ChatId,
		state:    state,
		sendTime: time.Now(),
	}
}

type HandlersCache interface {
	SaveState(context.Context, CachedInfo) error
	GetState(ctx context.Context, chatId int64) (*CachedInfo, error)
	SaveInfo(ctx context.Context, chatId int64, json string) error
	GetInfo(ctx context.Context, chatId int64) (string, error)
	AcquireLock(ctx context.Context, chatId int64, key string) *sync.Mutex
	ReleaseLock(ctx context.Context, chatId int64, key string)
	RemoveInfo(ctx context.Context, chatId int64) error
}
