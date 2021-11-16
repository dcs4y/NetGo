package ginterface

import (
	"context"
	"net"
	"time"
)

// IConnection 定义连接接口
type IConnection interface {
	Start()                   //启动连接，让当前连接开始工作
	Stop()                    //停止连接，结束当前连接状态M
	Context() context.Context //返回ctx，用于用户自定义的go程获取连接退出状态

	GetTCPConnection() *net.TCPConn //从当前连接获取原始的socket TCPConn
	GetConnID() uint64              //获取当前连接ID
	GetConnName() string            //获取当前连接名称
	RemoteAddr() net.Addr           //获取远程客户端地址信息

	SendMsg(message IMessage) error     //直接将Message数据发送数据给远程的TCP客户端(无缓冲)
	SendBuffMsg(message IMessage) error //直接将Message数据发送给远程的TCP客户端(有缓冲)

	SetCommandResponseAction(action string)            // 注册需要处理的应答指令
	HasCommandResponseAction(action string) bool       // 检测是否已注册需要处理的应答指令
	SetCommandResponse(message IMessage)               //设置下行指令同步应答的消息
	GetCommandResponse(timeout time.Duration) IMessage //获取下行指令同步应答的消息

	SetProperty(key string, value interface{})   //设置链接属性
	GetProperty(key string) (interface{}, error) //获取链接属性
	RemoveProperty(key string)                   //移除链接属性

	GetProtocolNo() string // 获取连接的协议
}
