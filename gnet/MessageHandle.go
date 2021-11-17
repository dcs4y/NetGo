package gnet

import (
	"fmt"
	"github.com/dcs4y/NetGo/ginterface"
)

// MessageHandle -
type MessageHandle struct {
	routers            map[string]ginterface.IRouter //存放每个MsgID 所对应的处理方法的map属性
	workerPoolSize     uint32                        //业务工作Worker池的数量
	maxWorkerQueueSize uint32                        //每个Worker的最大队列数
	requestQueue       []chan ginterface.IRequest    //Worker负责取任务的消息队列
}

//NewMessageHandle 创建MessageHandle
func NewMessageHandle(workerPoolSize, maxWorkerQueueSize uint32) *MessageHandle {
	return &MessageHandle{
		routers:            make(map[string]ginterface.IRouter),
		workerPoolSize:     workerPoolSize,
		maxWorkerQueueSize: maxWorkerQueueSize,
		//一个worker对应一个queue
		requestQueue: make([]chan ginterface.IRequest, workerPoolSize),
	}
}

//AddRouter 为消息添加具体的处理逻辑
//path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
func (mh *MessageHandle) AddRouter(path string, router ginterface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.routers[path]; ok {
		panic("repeated MessageHandle , path = " + path)
	}
	//2 添加msg与api的绑定关系
	mh.routers[path] = router
	fmt.Println("Add MessageHandle path = ", path)
}

//StartWorkerPool 启动worker工作池
func (mh *MessageHandle) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.workerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.requestQueue[i] = make(chan ginterface.IRequest, mh.maxWorkerQueueSize)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.requestQueue[i])
	}
}

//StartOneWorker 启动一个Worker工作流程
func (mh *MessageHandle) StartOneWorker(workerID int, requestQueue chan ginterface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-requestQueue:
			mh.doHandleRequest(request)
		}
	}
}

// HandleRequestToQueue 增加到HandleRequest的队列中，等待处理
func (mh *MessageHandle) HandleRequestToQueue(request ginterface.IRequest) {
	if mh.workerPoolSize > 0 {
		//已经启动工作池机制，将消息交给Worker处理
		//根据ConnID来分配当前的连接应该由哪个worker负责处理
		//轮询的平均分配法则

		//得到需要处理此条连接的workerID
		workerID := request.GetConnection().GetConnID() % uint64(mh.workerPoolSize)
		fmt.Printf("AddRequestToQueue ConnID=%d to workerID=%d, {ProtocolNo:%s,Action:%s,Body:%s}\n",
			request.GetConnection().GetConnID(),
			workerID,
			request.GetMessage().GetProtocolNo(), request.GetMessage().GetAction(), request.GetMessage().GetBody(),
		)
		//将请求消息发送给任务队列
		mh.requestQueue[workerID] <- request
	} else {
		//从绑定好的消息和对应的处理方法中执行对应的Handle方法
		go mh.doHandleRequest(request)
	}
}

//HandleRequest 马上以非阻塞方式处理消息
func (mh *MessageHandle) doHandleRequest(request ginterface.IRequest) {
	path := request.GetMessage().GetProtocolNo() + "_" + request.GetMessage().GetAction()
	handler, ok := mh.routers[path]
	if !ok {
		fmt.Println("MessageHandle path = ", path, " is not FOUND!")
		// 未注册路由的消息，默认为同步应答。将消息发送到同步应答通道等待处理。
		request.GetConnection().AddCommandResponse(request.GetMessage())
		return
	}
	//执行对应处理方法
	handler.Handle(request)
}

// HandleRequest 增加到HandleRequest的队列中，等待处理
func (mh *MessageHandle) HandleRequest(connection ginterface.IConnection, message ginterface.IMessage) {
	//得到当前客户端请求的Request数据
	request := &Request{
		connection: connection,
		message:    message,
	}
	mh.HandleRequestToQueue(request)
}
