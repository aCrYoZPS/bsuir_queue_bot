package interfaces

type CachedInfo struct {
	ChatId  int64
	Command string
}

func NewCachedInfo(ChatId int64, Command string) *CachedInfo {
	return &CachedInfo{
		ChatId:  ChatId,
		Command: Command,
	}
}

type HandlersCache interface {
	Save(CachedInfo) error
	Get(chatId int64) (*CachedInfo, error)
}
