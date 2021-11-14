# [NetGo](https://github.com/dcs4y/NetGo)

## 基于 [Zinx](https://github.com/aceld/zinx "Zinx主页") 的二次开发的TCP框架。支持GPS等自定义协议的扩展。

> ## 主要功能
>- 注册协议
>> 继承gnet.Connection，注册StartReader和Pack回调方法。
>- 注册路由
>> 继承ginterface.IRouter，实现Handle方法。
>- 支持指令下发

> 程序示例：
> ```
> func main() {
> 	s := gnet.NewServer("127.0.0.1", 8888, 1024, 10, 100)
> 	// 自定义协议示例
> 	s.OnNewConn(inet.NewConnection)
> 	// 路由添加示例 path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
> 	s.AddRouter("7878_01", &irouter.LoginRouter{})
> 	s.AddRouter("7878_13", &irouter.HeartbeatRouter{})
> 	// 指令下发示例
> 	c, b := s.GetConnectionManager().Get("")
> 	if b {
> 		c.SendMsg(gnet.NewMessage("", "", ""))
> 	}
> 	s.Start()
> }
> ```