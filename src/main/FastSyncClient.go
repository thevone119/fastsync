package main

import (
	"client"
	"zinx/zlog"
)
//同步的客户端，守护进程执行。同一份客户端只能运行一次，运行多次会有日志文件冲突等问题。
//如果需要运行多个客户端，需要拷贝到另外一个目录执行。

func main() {
	zlog.Info("FastSyncClient start...")
	//开启一个客户端监听处理
	c:=client.NewClient()
	c.Start()

	//阻塞主线程，永远不退出。
	select {}
}
