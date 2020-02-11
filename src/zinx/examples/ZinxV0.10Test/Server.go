package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/zlog"
	"zinx/znet"
)

//ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

//Ping Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	zlog.Info("Call HelloZinxRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(znet.NewMsgPackage(1, []byte("dddddd")))
	if err != nil {
		zlog.Error(err)
	}
}

type HelloZinxRouter struct {
	znet.BaseRouter
}

//HelloZinxRouter Handle
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")

	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendBuffMsg(znet.NewMsgPackage(1, []byte("dddddd")))
	if err != nil {
		zlog.Error(err)
	}
}

//创建连接的时候执行
func DoConnectionBegin(conn ziface.IConnection) {
	fmt.Println("DoConnecionBegin is Called ... ")

	//设置两个链接属性，在连接创建之后
	fmt.Println("Set conn Name, Home done!")
	conn.SetProperty("Name", "Aceld")
	conn.SetProperty("Home", "https://www.jianshu.com/u/35261429b7f1")

	err := conn.SendMsg(znet.NewMsgPackage(2, []byte("dddddd")))
	if err != nil {
		zlog.Error(err)
	}
}

//连接断开的时候执行
func DoConnectionLost(conn ziface.IConnection) {
	//在连接销毁之前，查询conn的Name，Home属性
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		fmt.Println("Conn Property Home = ", home)
	}

	fmt.Println("DoConneciotnLost is Called ... ")
}

func main() {
	//创建一个server句柄
	s := znet.NewServer()
	//zlog.SetLogFile("./log", "testfile2.log")
	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//配置路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})
	//开启服务
	s.Serve()

}
