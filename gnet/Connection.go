package gnet

import (
	"context"
	"errors"
	"fmt"
	"github.com/dcs4y/NetGo/ginterface"
	"net"
	"sync"
)

//Connection 链接
type Connection struct {
	//当前Conn属于哪个Server
	Server ginterface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint64
	// 当前连接名称，设备序列号 IMEI号
	ConnName string
	//消息管理MsgID和对应处理方法的消息管理模块
	MessageHandler ginterface.IMessageHandle
	//告知该链接已经退出/停止的channel
	Ctx    context.Context
	Cancel context.CancelFunc

	// StartReader 读消息Goroutine，用于从客户端中读取数据。由子类实现。
	StartReader func()

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	MessageChan chan []byte
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	MessageBuffChan chan []byte
	// 指令通道注册器。用于下行指令后，需要等待的处理结果。
	actionChan map[string]chan ginterface.IMessage

	// Pack 封包方法(压缩数据)。由具体的协议实现。
	Pack func(msg ginterface.IMessage) ([]byte, error)

	// 当前连接的锁
	sync.RWMutex
	//链接属性
	Property map[string]interface{}
	//保护当前property的锁
	PropertyLock sync.Mutex
	//当前连接的关闭状态
	IsClosed bool
}

//StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.MessageChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//fmt.Printf("Send data succ! data = %+v\n", data)
		case data, ok := <-c.MessageBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("MessageBuffChan is Closed")
				break
			}
		case <-c.Ctx.Done():
			return
		}
	}
}

//Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	c.Ctx, c.Cancel = context.WithCancel(context.Background())
	c.actionChan = make(map[string]chan ginterface.IMessage)
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.Server.CallOnConnStart(c)
}

//Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	c.Lock()
	defer c.Unlock()

	//如果当前链接已经关闭
	if c.IsClosed == true {
		return
	}

	fmt.Println("Conn Stop()...ConnID = ", c.ConnID)

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.Server.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()
	//关闭Writer
	c.Cancel()

	//将链接从连接管理器中删除
	c.Server.GetConnectionManager().Remove(c)

	//关闭该链接全部管道
	close(c.MessageChan)
	close(c.MessageBuffChan)
	//设置标志位
	c.IsClosed = true
}

//GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint64 {
	return c.ConnID
}

// GetConnName 获取当前连接名称
func (c *Connection) GetConnName() string {
	return c.ConnName
}

//RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(message ginterface.IMessage) error {
	c.RLock()
	defer c.RUnlock()
	if c.IsClosed == true {
		return errors.New("connection closed when send msg")
	}

	//将data封包，并且发送
	msg, err := c.Pack(message)
	if err != nil {
		fmt.Println(err)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.MessageChan <- msg

	return nil
}

//SendBuffMsg  发生BuffMsg
func (c *Connection) SendBuffMsg(message ginterface.IMessage) error {
	c.RLock()
	defer c.RUnlock()
	if c.IsClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	msg, err := c.Pack(message)
	if err != nil {
		fmt.Println(err)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.MessageBuffChan <- msg

	return nil
}

// RegisterCommandResponseChan 注册上行指令通道。调用者通过chan接收消息。
func (c *Connection) RegisterCommandResponseChan(action string, ch chan ginterface.IMessage) {
	c.actionChan[action] = ch
}

// AddCommandResponse 向注册的上行指令通道中添加消息。调用者通过chan接收消息。
func (c *Connection) AddCommandResponse(message ginterface.IMessage) {
	if ch, b := c.actionChan[message.GetAction()]; b {
		if len(ch) > 0 {
			fmt.Printf("AddCommandResponse delete message by %#v", message)
			// 删除旧的消息，防止阻塞
			fmt.Printf("AddCommandResponse delete message %#v", <-ch)
		} else {
			ch <- message
		}
	} else {
		fmt.Printf("AddCommandResponse no chan %#v", message)
	}
}

//SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.PropertyLock.Lock()
	defer c.PropertyLock.Unlock()
	if c.Property == nil {
		c.Property = make(map[string]interface{})
	}

	c.Property[key] = value
}

//GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.PropertyLock.Lock()
	defer c.PropertyLock.Unlock()

	if value, ok := c.Property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no Property found")
}

//RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.PropertyLock.Lock()
	defer c.PropertyLock.Unlock()
	delete(c.Property, key)
}

// Context 返回ctx，用于用户自定义的go程获取连接退出状态
func (c *Connection) Context() context.Context {
	return c.Ctx
}

// GetProtocolNo 获取连接的协议
func (c *Connection) GetProtocolNo() string {
	return ""
}
