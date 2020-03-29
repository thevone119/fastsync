package client

import (
	"comm"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zinx/zlog"
)

var ClientObj *Client

//客户端引擎
//1.实现客户端无限循环处理(200毫秒轮询)，这个其他类的一些需要定时循环处理的都加到这里
//2.其他组件，寻捆绑到这个客户端组件上，实现组件间的交互处理
type Client struct {
	Client *ClientUpManager
	NextClearTime int64	//下次清理日志的时间
	CurrUnixTime int64		//当前时间，秒
	NextAllSyncTime  int64	//下次全量同步时间，每分钟校验一次
	isRun bool
}

//创建一个新的客户端
func NewClient() *Client{
	return &Client{
		isRun:false,
		NextClearTime:0,
		NextAllSyncTime:0,
		CurrUnixTime:0,
	}
}

//开始
func (c *Client) Start() {
	zlog.Info("当前程序运行目录为:",comm.CURR_RUN_PATH,",程序日志将记录在此目录下")
	fmt.Println("当前程序运行目录为:",comm.CURR_RUN_PATH,",程序日志将记录在此目录下")
	//打开数据库
	comm.LeveldbDB.Open()
	//开启一个客户端监听处理
	c.Client = NewClientUpManager()
	//睡眠1秒等待网络连接
	time.Sleep(1 * time.Second)

	//开启文件监听
	for _, v := range comm.ClientConfigObj.NotifyMonitor {
		if !comm.ClientConfigObj.IsLocalPath(v){
			continue
		}
		fi,err:=os.Stat(v)
		if err!=nil || fi==nil||!fi.IsDir(){
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

		f := comm.NewPollWatch(comm.ClientConfigObj.PollMonitor)

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
		c.CurrUnixTime=time.Now().Unix()
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
	c.DoClearLog()
	c.DoAllSync()
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
		c.Client.ReSyncFile(e.Value.(*LocalFile))
	}
}

func (c *Client) DoFileChange(){
	l:=comm.FileChangeMonitorObj.GetQueue(200)
	for e := l.Front(); e != nil; e = e.Next() {
		if strings.Index(e.Value.(string),"del_")==0{
			c.Client.DeleteFile(e.Value.(string)[4:])
		}else{
			c.Client.SyncFile(e.Value.(string),comm.FCHECK_FULLMD5_CHECK,true,0)
		}
		//fmt.Print(e.Value) //输出list的值,01234
	}
}

//清理日志，10分钟一次
func (c *Client) DoClearLog(){
	//错误拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("DoClearLog 处理错误", p)
		}
	}()

	if c.CurrUnixTime<c.NextClearTime{
		return
	}
	c.NextClearTime=c.CurrUnixTime+60*10
	cday:=comm.ClientConfigObj.LogCleanDay
	if cday<1{
		cday=1
	}
	ctime:=time.Duration(cday*24)*time.Hour
	now:=time.Now()
	zlog.Info("clear log")
	//清理当前目录下的，超过X天的*.fstlog
	filepath.Walk(comm.CURR_RUN_PATH, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if m,_:=filepath.Match("*.fstlog",info.Name());m{
			if now.Sub(info.ModTime()) > ctime {
				os.Remove(path)
				zlog.Info("delete log",path)
				return nil
			}
		}
		return nil
	})
}


//自动同步，5秒一次，避免错过分钟，如果校验时间正确，则加2分钟，错过当前时间
func (c *Client) DoAllSync(){
	//错误拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("DoAllSync 处理错误", p)
		}
	}()
	if c.CurrUnixTime<c.NextAllSyncTime{
		return
	}
	c.NextAllSyncTime=c.CurrUnixTime+5

	//星期判断
	now:=time.Now()

	match:=false
	for _,v:=range comm.ClientConfigObj.AllSyncWeekday{
		if int(now.Weekday())==v{
			match=true
		}
	}
	if !match{
		return
	}
	match=false

	//判断时间符合
	for _,v:=range comm.ClientConfigObj.AllSyncTimeOfDay{
		if now.Format("15:04")==v{
			match=true
		}
	}
	if !match{
		return
	}
	//时间符合，则错开当前这个时间
	c.NextAllSyncTime=c.CurrUnixTime+70

	//调用外部
	zlog.Info("开始进行全量数据")
	//开一个携程进行全量同步哦
	go c.Client.SyncPath(comm.ClientConfigObj.AllSyncFileModTime,comm.BASE_PATH,comm.FCHECK_SIZE_AND_TIME_CHECK)

}
