package interfaces

import "github.com/google/uuid"

type GroupRequest struct {
	UUID   string
	MsgId  int64
	ChatId int64
}

func NewGroupRequest(msgId, chatId int64, opts ...func(*GroupRequest)) *GroupRequest {
	groupReq := &GroupRequest{MsgId: msgId, ChatId: chatId}
	for _, opt := range opts {
		opt(groupReq)
	}
	if groupReq.UUID == "" {
		groupReq.UUID = uuid.NewString()
	}
	return groupReq
}

func WithUUID(uuid string) func(req *GroupRequest) {
	return func(req *GroupRequest) {
		req.UUID = uuid
	}
}

type RequestsRepository interface {
	SaveRequest(*GroupRequest) error
	DeleteRequest(msgId int64) error
	GetByUUID(uuid string) ([]GroupRequest, error)
	GetByMsg(msgId, chatId int64) (*GroupRequest, error)
}
