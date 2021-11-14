package irouter

import (
	"fmt"
	"github.com/dcs4y/NetGo/example/iutil"
	"github.com/dcs4y/NetGo/ginterface"
	"github.com/dcs4y/NetGo/gnet"
)

type HeartbeatRouter struct {
}

func (router *HeartbeatRouter) Handle(request ginterface.IRequest) {
	oldMessage := request.GetMessage()
	fmt.Printf("%s=%s收到设备心跳消息 :%#v\n", iutil.GetFormatTime(), request.GetConnection().GetConnName(), oldMessage)
	// 响应客户端
	message := gnet.NewMessage(oldMessage.GetAction(), "", oldMessage.GetIndex())
	err := request.GetConnection().SendBuffMsg(message)
	if err != nil {
		fmt.Println(err)
	}
}
