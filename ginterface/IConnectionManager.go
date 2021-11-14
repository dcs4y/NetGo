package ginterface

// IConnectionManager 连接管理抽象层
type IConnectionManager interface {
	Add(conn IConnection)                    //添加链接
	Remove(conn IConnection)                 //删除连接
	Get(connName string) (IConnection, bool) //利用ConnName获取链接
	Length() int                             //获取当前的连接数
	ClearConn()                              //删除并停止所有链接
}
