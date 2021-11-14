package ginterface

// IRouter 路由接口， 这里面路由是 使用框架者给该链接自定的 处理业务方法路由里的IRequest，包含用该链接的链接信息和该链接的请求数据信息
type IRouter interface {
	Handle(request IRequest) //处理conn业务的方法
}
