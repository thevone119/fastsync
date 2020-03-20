package client

import (
	"comm"
	"fmt"
	"strings"
	"time"
	"zinx/zlog"
)

//客户端引擎
//1.实现客户端无限循环处理(200毫秒轮询)，这个其他类的一些需要定时循环处理的都加到这里
//2.其他组件，寻捆绑到这个客户端组件上，实现组件间的交互处理
type Client struct {
	client *ClientUpManager

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
	zlog.Info("当前程序运行目录为:",comm.CURR_RUN_PATH,",程序日志将记录在此目录下")
	fmt.Println("当前程序运行目录为:",comm.CURR_RUN_PATH,",程序日志将记录在此目录下")
	//打开数据库
	comm.LeveldbDB.Open()
	//开启一个客户端监听处理
	c.client = NewClientUpManager()
	//睡眠1秒等待网络连接
	time.Sleep(1 * time.Second)

	//开启文件监听
	for _, v := range comm.ClientConfigObj.NotifyMonitor {
		if !comm.ClientConfigObj.IsLocalPath(v){
			continue
		}
		fs := comm.NewFSWatch(v)
		fs.Start()
	}


	//监控文件
	for i, v := range comm.ClientConfigObj.LogMonitor {
		sep:=" "
		if len(comm.ClientConfigObj.LogMonitorSep)>i{
			sep=comm.ClientConfigObj.LogMonitorSep[i]
		}

		f := comm.NewLogWatch(v,sep)
		f.Start()
	}

	//轮训监控
	if len(comm.ClientConfigObj.PollMonitor)>0{
		f := comm.NewPollWatch(comm.ClientConfigObj.PollMonitor )
		f.Start()
	}





	//开启线程轮询
	c.isRun=true
	//开2个线程，一个是不耗时的，一个是耗时的
	go c.goDoHandle()

	go c.goDoUpload()

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
		time.Sleep(time.Millisecond*500)
	}
}

func (c *Client) goDoUpload(){
	for{
		if !c.isRun{
			return
		}
		c.goDoUpload2()
		time.Sleep(time.Millisecond*500)
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
func (c *Client) goDoUpload2(){
	//错误拦截,针对上传过程中遇到的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("DoUpload 处理错误", p)
		}
	}()
	c.DoUpload()
}


//处理，无限循环处理，携程调用
func (c *Client) DoHandle(){
	//
	//
	//
	LocalFileHandle.ClearTimeout()


}


//处理，无限循环处理，携程调用
func (c *Client) DoUpload(){
	//
	//
	//
	c.DoFileChange()
	c.DoReSendFile()
}

//重发失败的文件处理
func (c *Client) DoReSendFile(){
	l:=LocalFileHandle.GetReSend()
	for e := l.Front(); e != nil; e = e.Next() {
		c.client.ReSyncFile(e.Value.(*LocalFile))
	}
}

func (c *Client) DoFileChange(){
	l:=comm.FileChangeMonitorObj.GetQueue(200)
	for e := l.Front(); e != nil; e = e.Next() {
		if strings.Index(e.Value.(string),"del_")==0{
			c.client.DeleteFile(e.Value.(string)[4:])
		}else{
			c.client.SyncFile(e.Value.(string),comm.CheckFileType(3),true)
		}
		//fmt.Print(e.Value) //输出list的值,01234
	}
}
