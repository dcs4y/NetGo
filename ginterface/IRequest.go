package ginterface

// IRequest 实际上是把客户端请求的链接信息 和 请求的数据 包装到了Request里
type IRequest interface {
	GetConnection() IConnection //获取请求连接
	GetMessage() IMessage       //获取请求消息的数据

	ResetConnection(connection IConnection) //重置下发的连接
	GetConnName() string                    //获取下发消息时的连接名称
}
