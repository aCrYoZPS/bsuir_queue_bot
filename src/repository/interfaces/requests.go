package interfaces

type Request struct {
	msgId  int64
	chatId int64
}

func NewRequest(msgId, chatId int64) *Request {
	return &Request{msgId: msgId, chatId: chatId}
}

func (req *Request) MsgID() int64 {
	return req.msgId
}

func (req *Request) ChatId() int64 {
	return req.chatId
}

type RequestsRepository interface {
	SaveRequest(*Request) error
	DeleteRequest(*Request) error
}

