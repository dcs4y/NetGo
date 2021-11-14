package gnet

import (
	"fmt"
	"github.com/dcs4y/NetGo/ginterface"
	"net"
)

var banner = `

     NNNNNNNN        NNNNNNNN                             tttt                 GGGGGGGGGGGGG
     N:::::::N       N::::::N                          ttt:::t              GGG::::::::::::G
     N::::::::N      N::::::N                          t:::::t            GG:::::::::::::::G                       █▀▀▀▀▀▀▀███▀███▀███▀▀▀▀▀▀▀█
     N:::::::::N     N::::::N                          t:::::t           G:::::GGGGGGGG::::G                       █ █▀▀▀█ ██  ▀██ █ █ █▀▀▀█ █
     N::::::::::N    N::::::N    eeeeeeeeeeee    ttttttt:::::ttttttt    G:::::G       GGGGGG   ooooooooooo         █ █   █ █ █▄█▄ ▄▄ █ █   █ █
     N:::::::::::N   N::::::N  ee::::::::::::ee  t:::::::::::::::::t   G:::::G               oo:::::::::::oo       █ ▀▀▀▀▀ █▀▄ █ ▄ █ █ ▀▀▀▀▀ █
     N:::::::N::::N  N::::::N e::::::eeeee:::::eet:::::::::::::::::t   G:::::G              o:::::::::::::::o      █▀▀███▀▀▀█ ▀██▀  ████▀▀████
     N::::::N N::::N N::::::Ne::::::e     e:::::etttttt:::::::tttttt   G:::::G    GGGGGGGGGGo:::::ooooo:::::o      █ █   ▄▀▀█▀▄▄██  ███   ▄ ▀█
     N::::::N  N::::N:::::::Ne:::::::eeeee::::::e      t:::::t         G:::::G    G::::::::Go::::o     o::::o      █▀█▀ ▄▄▀ ▄▀ ▀▀██▄ ▄  █▄██ █
     N::::::N   N:::::::::::Ne:::::::::::::::::e       t:::::t         G:::::G    GGGGG::::Go::::o     o::::o      █ ██▀ █▀ ▄  ▄▀▀█ █ ▀ ▀▀█ ▀█
     N::::::N    N::::::::::Ne::::::eeeeeeeeeee        t:::::t         G:::::G        G::::Go::::o     o::::o      █ █▄▀█▀▀  ▄  ▀▀▀  ▀  ▀▄ █▄█
     N::::::N     N:::::::::Ne:::::::e                 t:::::t    ttttttG:::::G       G::::Go::::o     o::::o      █▀▀▀▀▀▀▀█ █ ▀██ █ █▀█ ███▀█
     N::::::N      N::::::::Ne::::::::e                t::::::tttt:::::t G:::::GGGGGGGG::::Go:::::ooooo:::::o      █ █▀▀▀█ █▄█  ▀▀▄  ▀▀▀ █▀▀██
     N::::::N       N:::::::N e::::::::eeeeeeee        tt::::::::::::::t  GG:::::::::::::::Go:::::::::::::::o      █ █   █ ██▄▀▄▄█▀▄▄▄▄██▀▀▄ █
     N::::::N        N::::::N  ee:::::::::::::e          tt:::::::::::tt    GGG::::::GGG:::G oo:::::::::::oo       █ ▀▀▀▀▀ █ ▀▀█▄██▄ ▀▀▄▄▀██ █
     NNNNNNNN         NNNNNNN    eeeeeeeeeeeeee            ttttttttttt         GGGGGG   GGGG   ooooooooooo         ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀

`

//Server 接口实现，定义一个Server服务类
type Server struct {
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port    int
	MaxConn int

	//当前Server的消息管理模块，用来绑定MsgID和对应的处理方法
	messageHandler ginterface.IMessageHandle
	//当前Server的链接管理器
	connectionManager ginterface.IConnectionManager
	//当前服务所接收到的连接计数。从1开始。
	connectionIndex uint64

	//注册每个真实连接创建时的协议解析方法
	onNewConn func(server ginterface.IServer, conn *net.TCPConn) (ginterface.IConnection, error)

	//该Server的连接创建时Hook函数
	onConnStart func(conn ginterface.IConnection)
	//该Server的连接断开时的Hook函数
	onConnStop func(conn ginterface.IConnection)
}

//NewServer 创建一个服务器句柄
func NewServer(ip string, port, maxConnectionNo int, workerPoolSize, maxWorkerQueueSize uint32) ginterface.IServer {
	fmt.Print(banner)
	s := &Server{
		IPVersion:         "tcp4",
		IP:                ip,
		Port:              port,
		MaxConn:           maxConnectionNo,
		messageHandler:    NewMessageHandle(workerPoolSize, maxWorkerQueueSize),
		connectionManager: NewConnectionManager(),
	}
	return s
}

//============== 实现 ginterface.IServer 里的全部接口方法 ========

//Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] NetGo Server listening at IP: %s, Port %d is starting\n", s.IP, s.Port)

	//开启一个go去做服务端Listener业务
	go func() {
		//0 启动worker工作池机制
		s.messageHandler.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		//已经监听成功
		fmt.Println("start NetGo server  success, now listening...")

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}
			fmt.Println("Get conn remote addr = ", conn.RemoteAddr().String())

			s.connectionIndex++

			//3.2 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.connectionManager.Length() >= s.MaxConn {
				conn.Close()
				continue
			}

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			connection, err := s.onNewConn(s, conn)
			if err != nil {
				fmt.Println("Create Connection err ", err)
				continue
			}
			//将新创建的Conn添加到链接管理中
			s.GetConnectionManager().Add(connection)

			//3.4 启动当前链接的处理业务
			go connection.Start()
		}
	}()
	//阻塞,否则主Go退出，listener的go将会退出
	select {}
}

//Stop 停止服务
func (s *Server) Stop() {
	fmt.Println("[STOP] NetGo Server！")
	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.connectionManager.ClearConn()
}

//AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
//path := IMessage.GetProtocolNo() + "_" + IMessage.GetAction()
func (s *Server) AddRouter(path string, router ginterface.IRouter) {
	s.messageHandler.AddRouter(path, router)
}

func (s *Server) GetMessageHandle() ginterface.IMessageHandle {
	return s.messageHandler
}

//GetConnectionManager 得到链接管理
func (s *Server) GetConnectionManager() ginterface.IConnectionManager {
	return s.connectionManager
}

// GetConnectionIndex 获取当前服务所接收到的连接数
func (s *Server) GetConnectionIndex() uint64 {
	return s.connectionIndex
}

func (s *Server) OnNewConn(onNewConn func(server ginterface.IServer, conn *net.TCPConn) (ginterface.IConnection, error)) {
	s.onNewConn = onNewConn
}

//OnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) OnConnStart(hookFunc func(ginterface.IConnection)) {
	s.onConnStart = hookFunc
}

//OnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) OnConnStop(hookFunc func(ginterface.IConnection)) {
	s.onConnStop = hookFunc
}

//CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ginterface.IConnection) {
	if s.onConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.onConnStart(conn)
	}
}

//CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ginterface.IConnection) {
	if s.onConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.onConnStop(conn)
	}
}
