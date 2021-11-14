package gnet

import (
	"fmt"
	"github.com/dcs4y/NetGo/ginterface"
	"sync"
)

//ConnectionManager 连接管理模块
type ConnectionManager struct {
	connections map[string]ginterface.IConnection //管理的连接信息
	connLock    sync.RWMutex                      //读写连接的读写锁
}

//NewConnectionManager 创建一个链接管理
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]ginterface.IConnection),
	}
}

//Add 添加链接
func (cm *ConnectionManager) Add(conn ginterface.IConnection) {
	//保护共享资源Map 加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//将conn连接添加到ConnMananger中
	cm.connections[conn.GetConnName()] = conn

	fmt.Println("connection add to ConnectionManager successfully: conn num = ", cm.Length())
}

//Remove 删除连接
func (cm *ConnectionManager) Remove(conn ginterface.IConnection) {
	//保护共享资源Map 加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//删除连接信息
	delete(cm.connections, conn.GetConnName())

	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", cm.Length())
}

//Get 利用ConnName获取链接
func (cm *ConnectionManager) Get(connName string) (ginterface.IConnection, bool) {
	//保护共享资源Map 加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	conn, ok := cm.connections[connName]
	return conn, ok
}

//Length 获取当前的连接数
func (cm *ConnectionManager) Length() int {
	return len(cm.connections)
}

//ClearConn 清除并停止所有连接
func (cm *ConnectionManager) ClearConn() {
	//保护共享资源Map 加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range cm.connections {
		//停止
		conn.Stop()
		//删除
		delete(cm.connections, connID)
	}

	fmt.Println("Clear All Connections successfully: conn num = ", cm.Length())
}

//ClearOneConn  利用ConnName获取一个链接 并且删除
func (cm *ConnectionManager) ClearOneConn(connName string) {
	//保护共享资源Map 加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	if conn, ok := cm.connections[connName]; !ok {
		//停止
		conn.Stop()
		//删除
		delete(cm.connections, connName)
		fmt.Println("Clear Connections Name:  ", connName, "succeed")
		return
	}

	fmt.Println("Clear Connections Name:  ", connName, "err")
	return
}
