package interfaces

import "context"

type AdminRequest struct {
	UUID   string
	MsgId  int64
	ChatId int64
}

func NewAdminRequest(msgId, chatId int64, uuid string) *AdminRequest {
	return &AdminRequest{MsgId: msgId, ChatId: chatId, UUID: uuid}
}

type AdminRequestsRepository interface {
	SaveRequest(ctx context.Context, req *AdminRequest) error
	DeleteRequest(ctx context.Context, msgId int64) error
	GetByUUID(ctx context.Context, uuid string) ([]AdminRequest, error)
	GetByMsg(ctx context.Context, msgId, chatId int64) (*AdminRequest, error)
}
