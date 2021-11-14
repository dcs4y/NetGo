package ginterface

import (
	"net"
)

// IServer 定义服务接口
type IServer interface {
	Start() //启动服务器方法
	Stop()  //停止服务器方法

	// AddRouter path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
	AddRouter(path string, router IRouter) //路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用

	GetMessageHandle() IMessageHandle         //得到消息处理器
	GetConnectionManager() IConnectionManager //得到链接管理器
	GetConnectionIndex() uint64               //获取当前服务所接收到的连接数

	OnNewConn(func(Server IServer, conn *net.TCPConn) (IConnection, error)) //注册每个真实连接创建时的协议解析方法

	OnConnStart(func(IConnection))    //注册每个连接创建后的Hook函数
	OnConnStop(func(IConnection))     //注册每个连接断开前的Hook函数
	CallOnConnStart(conn IConnection) //调用连接OnConnStart Hook函数
	CallOnConnStop(conn IConnection)  //调用连接OnConnStop Hook函数
}
