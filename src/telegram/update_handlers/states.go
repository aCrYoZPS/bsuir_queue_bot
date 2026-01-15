package update_handlers

import (
	"context"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type Cache interface {
	SaveState(context.Context, interfaces.CachedInfo) error
	GetState(ctx context.Context, chatId int64) (*interfaces.CachedInfo, error)
	SaveInfo(ctx context.Context, chatId int64, json string) error
	GetInfo(ctx context.Context, chatId int64) (string, error)
	AcquireLock(ctx context.Context, chatId int64, key string) *sync.Mutex
	ReleaseLock(ctx context.Context, chatId int64, key string)
	RemoveInfo(ctx context.Context, chatId int64) error
}
