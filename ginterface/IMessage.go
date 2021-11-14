package ginterface

// IMessage 将请求的一个消息封装到message中，定义抽象层接口
type IMessage interface {
	GetProtocolNo() string // 协议号
	GetDataLen() string    // 包长度
	GetAction() string     // 事件
	GetBody() string       // 信息内容
	GetBodyData() []byte   // 信息内容
	GetIndex() string      // 信息序列号

	SetProtocolNo(protocolNo string) // 协议号
}
