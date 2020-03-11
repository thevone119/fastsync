package client

import (
	"comm"
	"time"
	"zinx/zlog"
)

//客户端引擎
//1.实现客户端无限循环处理(200毫秒轮询)，这个其他类的一些需要定时循环处理的都加到这里
//2.其他组件，寻捆绑到这个客户端组件上，实现组件间的交互处理

type Client struct {
	client *ClientUpManager
	fswatch *comm.FSWatch
	fchan *comm.FileChan

	isRun bool
}

//创建一个新的客户端
func NewClient() *Client{
	return &Client{
		isRun:false,
	}
}

//开始
func (c *Client) Start() {
	//打开数据库
	comm.LeveldbDB.Open()
	//开启一个客户端监听处理
	c.client = NewClientUpManager()
	//睡眠1秒等待网络连接
	time.Sleep(1 * time.Second)
	//开启文件监听
	fs := comm.NewFSWatch("d:/video")
	fs.Start()
	//监控文件
	f := comm.NewFileChan(comm.NOTIFY_PATH, "*.nlog")
	f.Start()
	//开启线程轮询
	c.isRun=true
	go c.goDoHandle()
}

//结束
func (c *Client) Stop() {
	c.isRun=false

}


//处理，无限循环处理，携程调用
func (c *Client) goDoHandle(){
	for{
		if !c.isRun{
			return
		}
		c.goDoHandle2()
		time.Sleep(time.Millisecond*200)
	}
}

//处理，无限循环处理，携程调用
func (c *Client) goDoHandle2(){
	//错误拦截,针对上传过程中遇到的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("goDoHandle 处理错误", p)
		}
	}()
	c.DoHandle()
}


//处理，无限循环处理，携程调用
func (c *Client) DoHandle(){

}