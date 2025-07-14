package memory

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

// TODO: create background processes to clean the maps for GC
type HandlersCache struct {
	interfaces.HandlersCache
	storage sync.Map
	locks   sync.Map
}

func NewHandlersCache() *HandlersCache {
	return &HandlersCache{
		storage: sync.Map{},
	}
}

func (cache *HandlersCache) Save(info interfaces.CachedInfo) error {
	cache.storage.Store(info.ChatId(), info)
	return nil
}

func (cache *HandlersCache) Get(chatId int64) (*interfaces.CachedInfo, error) {
	value, ok := cache.storage.LoadAndDelete(chatId)
	if !ok {
		return nil, errors.New("no info cached")
	}
	cached, ok := value.(interfaces.CachedInfo)
	if !ok {
		return nil, errors.New("info cached is in incorrect type")
	}
	return &cached, nil
}

func (cache *HandlersCache) AcquireLock(chatId int64) *sync.Mutex {
	mu := &sync.Mutex{}
	val, loaded := cache.locks.LoadOrStore(chatId, mu)
	if loaded {
		mu, _ = val.(*sync.Mutex)
	}
	return mu
}

func (cache *HandlersCache) ReleaseLock(chatId int64) {
	cache.locks.Delete(chatId)
}
