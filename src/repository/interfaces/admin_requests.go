package interfaces

type AdminRequest struct {
	UUID   string
	MsgId  int64
	ChatId int64
}

func NewAdminRequest(msgId, chatId int64, uuid string) *AdminRequest {
	return &AdminRequest{MsgId: msgId, ChatId: chatId, UUID: uuid}
}

type AdminRequestsRepository interface {
	SaveRequest(*AdminRequest) error
	DeleteRequest(msgId int64) error
	GetByUUID(uuid string) ([]AdminRequest, error)
	GetByMsg(msgId, chatId int64) (*AdminRequest, error)
}
