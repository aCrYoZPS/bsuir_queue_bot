package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/aCrYoZPS/bsuir_queue_bot/src/repository/interfaces"
)

// TODO: create background processes to clean the maps for GC

var _ interfaces.HandlersCache = (*HandlersCache)(nil)
type HandlersCache struct {
	db    *sql.DB
	locks sync.Map
}

func NewHandlersCache(db *sql.DB) *HandlersCache {
	return &HandlersCache{
		db:    db,
		locks: sync.Map{},
	}
}

const (
	STATES_TABLE = "states"
	INFO_TABLE   = "info"
)

func (cache *HandlersCache) SaveState(ctx context.Context, info interfaces.CachedInfo) error {
	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (chat_id, state) VALUES ($1, $2)", STATES_TABLE)
	_, err := cache.db.ExecContext(ctx, query, info.ChatId(), info.State())
	return err
}

func (cache *HandlersCache) GetState(ctx context.Context, chatId int64) (*interfaces.CachedInfo, error) {
	query := fmt.Sprintf("SELECT state FROM %s WHERE chat_id=$1", STATES_TABLE)
	row := cache.db.QueryRowContext(ctx, query, chatId)
	if row.Err() != nil {
		return nil, row.Err()
	}
	state := ""
	err := row.Scan(&state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return interfaces.NewCachedInfo(chatId, state), nil
		}
		return nil, err
	}
	return interfaces.NewCachedInfo(chatId, state), nil
}

func (cache *HandlersCache) AcquireLock(ctx context.Context, chatId int64) *sync.Mutex {
	mu := &sync.Mutex{}
	val, loaded := cache.locks.LoadOrStore(chatId, mu)
	if loaded {
		mu, _ = val.(*sync.Mutex)
	}
	return mu
}

func (cache *HandlersCache) ReleaseLock(ctx context.Context, chatId int64) {
	cache.locks.Delete(chatId)
}

func (cache *HandlersCache) SaveInfo(ctx context.Context, chatId int64, json string) error {
	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (chat_id, json) VALUES ($1, $2)", INFO_TABLE)
	_, err := cache.db.ExecContext(ctx, query, chatId, json)
	return err
}

func (cache *HandlersCache) GetInfo(ctx context.Context, chatId int64) (string, error) {
	query := fmt.Sprintf("SELECT json FROM %s WHERE chat_id=$1", INFO_TABLE)
	row := cache.db.QueryRowContext(ctx, query, chatId)
	if row.Err() != nil {
		return "", row.Err()
	}
	json := ""
	err := row.Scan(&json)
	if err != nil {
		return "", err
	}
	return json, nil
}
