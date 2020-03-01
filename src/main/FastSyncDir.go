package main

import (
	"client"
	"comm"
	"flag"
	"time"
	"zinx/zlog"
)

//同步目录
//接受外部参数 -t 0 -p /test -c 3
//-t 时间（分钟）
//-p 路径
//-c 校验类型
//执行完后直接退出
func main() {
	// 定义几个变量，用于接收命令行的参数值
	var ltime int64
	var lpath string
	var filecheck int
	// &user 就是接收命令行中输入 -u 后面的参数值，其他同理
	flag.Int64Var(&ltime, "t", 0, "文件最后修改时间，默认0")
	flag.StringVar(&lpath, "p", "/", "同步文件路径，默认根路径")
	flag.IntVar(&filecheck, "c", 0, "同步文件校验类型，默认3(完整的MD5校验)")
	// 解析命令行参数写入注册的flag里
	flag.Parse()
	zlog.Info("开始执行全量文件同步，同步时间:", ltime, "同步路径:", "文件校验类型:", filecheck)
	//计算耗时，XX毫秒
	currTime := time.Now().UnixNano() / 1e6
	start(ltime, lpath, comm.CheckFileType(1))
	zlog.Info("全量文件同步文件同步执行完成，耗时:", time.Now().UnixNano()/1e6-currTime, "毫秒")
}

func start(ltime int64, lpath string, filecheck comm.CheckFileType) {
	c := client.NewClientUpManager()

	for {
		//c.SyncPath("D:/code")
		c.SyncPath("e:/project/")
		time.Sleep(time.Second * 60 * 5)
	}

}