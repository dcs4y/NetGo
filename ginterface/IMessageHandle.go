package ginterface

// IMessageHandle 消息管理抽象层
type IMessageHandle interface {
	// AddRouter path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
	AddRouter(path string, router IRouter)                  //为消息添加具体的处理逻辑
	StartWorkerPool()                                       //启动worker工作池
	HandleRequestToQueue(request IRequest)                  //增加到HandleRequest的队列中，等待处理
	HandleRequest(connection IConnection, message IMessage) //增加到HandleRequest的队列中，等待处理
}
