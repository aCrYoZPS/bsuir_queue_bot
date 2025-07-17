package memory

import (
	"errors"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

// TODO: create background processes to clean the maps for GC
type HandlersCache struct {
	interfaces.HandlersCache
	stateStorage sync.Map
	infoStorage  sync.Map
	locks        sync.Map
}

func NewHandlersCache() *HandlersCache {
	return &HandlersCache{
		stateStorage: sync.Map{},
		infoStorage:  sync.Map{},
		locks:        sync.Map{},
	}
}

func (cache *HandlersCache) SaveState(info interfaces.CachedInfo) error {
	cache.stateStorage.Store(info.ChatId(), info)
	return nil
}

func (cache *HandlersCache) GetState(chatId int64) (*interfaces.CachedInfo, error) {
	value, ok := cache.stateStorage.LoadAndDelete(chatId)
	if !ok {
		return nil, nil
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

func (cache *HandlersCache) SaveInfo(chatId int64, json string) error {
	cache.infoStorage.Store(chatId, json)
	return nil
}

func (cache *HandlersCache) GetInfo(chatId int64) (string, error) {
	val, ok := cache.infoStorage.Load(chatId)
	if !ok {
		return "", nil
	}
	info, ok := val.(string)
	if !ok {
		return "", errors.New("info cached is in incorrect type")
	}
	return info, nil
}
func (cache *HandlersCache) RemoveInfo(chatId int64) error {
	cache.infoStorage.Delete(chatId)
	return nil
}
