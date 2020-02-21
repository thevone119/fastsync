package main

import (
	"comm"
	"fmt"
	"server"
	"zinx/ziface"
	"zinx/znet"
)

//创建连接的时候执行
func DoConnectionBegin(conn ziface.IConnection) {
	fmt.Println("DoConnecionBegin is Called ... ")
}

//连接断开的时候执行
func DoConnectionLost(conn ziface.IConnection) {
	fmt.Println("DoConnectionLost is Called ... ")
	//在连接销毁之前，做连接捆绑内容的清理
	server.SyncFileHandle.CloseAll(conn.GetConnID())
}

func main() {
	//创建一个server句柄
	s := znet.NewServer()
	//zlog.SetLogFile("./log", "testfile2.log")
	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//配置路由,保持连接
	s.AddRouter(comm.MID_KeepAlive, &server.KeepAliveRouter{})
	//配置路由,登录
	s.AddRouter(comm.MID_Login, &server.LoginRouter{})
	//配置路由,校验文件
	s.AddRouter(comm.MID_CheckFile, &server.CheckFileRouter{})
	//配置路由,发送文件上传请求
	s.AddRouter(comm.MID_SendFileReq, &server.SendFileReqRouter{})
	//配置路由,发送请求
	s.AddRouter(comm.MID_Request, &server.RequestRouter{})
	//配置路由,发送文件
	s.AddRouter(comm.MID_SendFile, &server.SendFileMsgRouter{})
	//配置路由,删除文件、文件夹
	s.AddRouter(comm.MID_DeleteFileReq, &server.DeleteFileRouter{})
	//配置路由,复制文件，移动文件，文件夹
	s.AddRouter(comm.MID_MoveFileReq, &server.MoveFileRouter{})

	//开启服务
	s.Serve()

}
