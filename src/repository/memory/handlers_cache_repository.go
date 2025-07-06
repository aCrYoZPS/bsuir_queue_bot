package memory

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

type HandlersCache struct {
	interfaces.HandlersCache
	storage sync.Map
}

func NewHandlersCache() *HandlersCache {
	return &HandlersCache{
		storage: sync.Map{},
	}
}

func (cache *HandlersCache) Save(info interfaces.CachedInfo) error {
	cache.storage.Store(info.ChatId, info)
	return nil
}

func (cache *HandlersCache) Get(chatId int64) (*interfaces.CachedInfo, error) {
	value, ok := cache.storage.LoadAndDelete(chatId)
	if !ok {
		return nil, errors.New("No info cached")
	}
	cached, ok := value.(interfaces.CachedInfo)
	if !ok {
		return nil, errors.New("Info cached is in incorrect type")
	}
	return &cached, nil
}
