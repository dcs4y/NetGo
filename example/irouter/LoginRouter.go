package irouter

import (
	"fmt"
	"github.com/dcs4y/NetGo/example/iutil"
	"github.com/dcs4y/NetGo/ginterface"
	"github.com/dcs4y/NetGo/gnet"
)

type LoginRouter struct {
}

func (router *LoginRouter) Handle(request ginterface.IRequest) {
	oldMessage := request.GetMessage()
	fmt.Printf("%s=%s收到设备登录消息 :%#v\n", iutil.GetFormatTime(), request.GetConnection().GetConnName(), oldMessage)
	// 响应客户端
	message := gnet.NewMessage(oldMessage.GetAction(), "", oldMessage.GetIndex())
	err := request.GetConnection().SendBuffMsg(message)
	if err != nil {
		fmt.Println(err)
	}
}
