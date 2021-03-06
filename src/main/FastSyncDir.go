package main

import (
	"client"
	"comm"
	"flag"
	"os"
	"path/filepath"
	"time"
	"zinx/zlog"
)
var IS_EXIT=false
//同步目录
//接受外部参数 -t 0 -p /test -c 3
//-t 时间（秒）
//-p 路径 /
//-c 校验类型
//-l 是否记录同步文件日志，默认不记录
//执行完后直接退出
func main() {
	// 定义几个变量，用于接收命令行的参数值
	var ltime int64
	var lpath string
	var filecheck int
	var islog int
	// &user 就是接收命令行中输入 -u 后面的参数值，其他同理
	flag.Int64Var(&ltime, "t", 0, "文件最后修改时间，默认0")
	flag.StringVar(&lpath, "p", "/", "同步文件路径，默认根路径")
	flag.IntVar(&filecheck, "c", 3, "同步文件校验类型，默认3(完整的MD5校验)")
	flag.IntVar(&islog, "l", 1, "是否记录同步文件日志，默认1(记录)")
	// 解析命令行参数写入注册的flag里
	flag.Parse()
	zlog.Info("开始执行全量文件同步，同步时间:", ltime, "秒，同步路径:", lpath,"文件校验类型:", filecheck,"是否记录同步文件日志:",islog)
	zlog.Info("当前程序运行目录为:",comm.CURR_RUN_PATH,",程序日志将记录在此目录下")
	if islog==0{
		client.LocalFileHandle.IsLog=false
	}else{
		client.LocalFileHandle.IsLog=true
	}
	t1 := time.Now()
	lpath=filepath.Join(comm.BASE_PATH,lpath)
	fp,err:=os.Stat(lpath)
	if err!=nil{
		zlog.Error("同步出错，找不到此路径",lpath,err)
		return
	}
	if !fp.IsDir(){
		zlog.Error("同步出错，此路径非目录",lpath)
		return
	}
	//携程处理一些任务
	go goDoHandle()
	//
	c := client.NewClientUpManager()
	c.SyncPath(ltime, lpath, comm.CheckFileType(filecheck))
	//阻塞等待所有上传文件结束
	client.SyncFileWG.Wait()
	IS_EXIT=true
	zlog.Info("文件同步执行完成，耗时:", time.Now().Sub(t1),"同步文件数:",client.LocalFileHandle.UpLoadFileCount,"成功同步:",client.LocalFileHandle.SuccUpLoadCount,"失败同步:",client.LocalFileHandle.ErrUpLoadCount)
}


func goDoHandle(){
	for{
		if IS_EXIT{
			return
		}
		goDoHandle2()

		time.Sleep(time.Second)
	}
}

func goDoHandle2(){
	//错误拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("goDoHandle2 处理错误", p)
		}
	}()
	client.LocalFileHandle.ClearTimeout()
}
