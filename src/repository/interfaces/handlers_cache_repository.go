package interfaces

import (
	"context"
	"sync"
	"time"
)

type CachedInfo struct {
	chatId   int64
	state    string
	sendTime time.Time
}

func (info *CachedInfo) State() string {
	return info.state
}

func (info *CachedInfo) SendTime() time.Time {
	return info.sendTime
}

func (info *CachedInfo) ChatId() int64 {
	return info.chatId
}

func NewCachedInfo(ChatId int64, State string) *CachedInfo {
	return &CachedInfo{
		chatId:   ChatId,
		state:    State,
		sendTime: time.Now().UTC(),
	}
}

type HandlersCache interface {
	SaveState(context.Context, CachedInfo) error
	GetState(ctx context.Context, chatId int64) (*CachedInfo, error)
	SaveInfo(ctx context.Context, chatId int64, json string) error
	GetInfo(ctx context.Context, chatId int64) (string, error)
	AcquireLock(ctx context.Context, chatId int64) *sync.Mutex
	ReleaseLock(ctx context.Context, chatId int64)
	RemoveInfo(ctx context.Context, chatId int64) error
}
