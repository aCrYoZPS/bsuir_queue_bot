package interfaces

import "time"

type CachedInfo struct {
	chatId   int64
	command  string
	sendTime time.Time
}

func (info *CachedInfo) Command() string {
	return info.command
}

func (info *CachedInfo) SendTime() time.Time {
	return info.sendTime
}

func (info *CachedInfo) ChatId() int64 {
	return info.chatId
}

func NewCachedInfo(ChatId int64, Command string) *CachedInfo {
	return &CachedInfo{
		chatId:   ChatId,
		command:  Command,
		sendTime: time.Now().UTC(),
	}
}

type HandlersCache interface {
	Save(CachedInfo) error
	Get(chatId int64) (*CachedInfo, error)
}
