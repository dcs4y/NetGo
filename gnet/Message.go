package gnet

import (
	"strconv"
)

//Message 消息
type Message struct {
	ProtocolNo string // 协议号
	DataLen    string // 包长度
	Action     string // 事件
	Body       string // 信息内容
	BodyData   []byte // 信息内容
	Index      string // 信息序列号(16进制)
}

//NewMessage 创建一个Message消息包
func NewMessage(action string, body string, index string) *Message {
	length := len(action+body+index)/2 + 2
	dataLen := strconv.FormatInt(int64(length), 16)
	if len(dataLen) == 1 {
		dataLen = "0" + dataLen
	}
	return &Message{
		DataLen: dataLen,
		Action:  action,
		Body:    body,
		Index:   index,
	}
}

// GetProtocolNo 协议号
func (msg *Message) GetProtocolNo() string {
	return msg.ProtocolNo
}

// GetDataLen 包长度
func (msg *Message) GetDataLen() string {
	return msg.DataLen
}

// GetAction 事件
func (msg *Message) GetAction() string {
	return msg.Action
}

// GetBody 信息内容
func (msg *Message) GetBody() string {
	return msg.Body
}

// GetBodyData 信息内容
func (msg *Message) GetBodyData() []byte {
	return msg.BodyData
}

// GetIndex 信息序列号
func (msg *Message) GetIndex() string {
	return msg.Index
}

// SetProtocolNo 协议号
func (msg *Message) SetProtocolNo(protocolNo string) {
	msg.ProtocolNo = protocolNo
}
