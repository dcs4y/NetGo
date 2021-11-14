package gnet

import (
	"github.com/dcs4y/NetGo/ginterface"
)

//Request 请求
type Request struct {
	connection     ginterface.IConnection //已经和客户端建立好的链接
	connectionName string                 //下发消息时的连接名称
	message        ginterface.IMessage    //客户端请求的数据
}

//GetConnection 获取请求连接信息
func (r *Request) GetConnection() ginterface.IConnection {
	return r.connection
}

// GetMessage 获取请求消息的数据
func (r *Request) GetMessage() ginterface.IMessage {
	return r.message
}

// ResetConnection 重置下发的连接
func (r *Request) ResetConnection(connection ginterface.IConnection) {
	r.connection = connection
}

// GetConnName 获取下发消息时的连接名称
func (r *Request) GetConnName() string {
	return r.connectionName
}
