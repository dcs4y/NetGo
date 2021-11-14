package main

import (
	"github.com/dcs4y/NetGo/example/inet"
	"github.com/dcs4y/NetGo/example/irouter"
	"github.com/dcs4y/NetGo/gnet"
)

func main() {
	s := gnet.NewServer("127.0.0.1", 8888, 1024, 10, 100)
	// 自定义协议示例
	s.OnNewConn(inet.NewConnection)
	// 路由添加示例 path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
	s.AddRouter("7878_01", &irouter.LoginRouter{})
	s.AddRouter("7878_13", &irouter.HeartbeatRouter{})
	// 消息发送示例
	c, b := s.GetConnectionManager().Get("")
	if b {
		c.SendMsg(gnet.NewMessage("", "", ""))
	}
	s.Start()
}
