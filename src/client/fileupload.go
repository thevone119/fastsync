package client

import (
	"comm"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"utils"
	"zinx/ziface"
)

//文件上传类
type FileUpload struct {
	netclient *NetWork
	timeout int64	//超时时间，秒

	sendMapMutex sync.RWMutex	//所有map操作的读写锁
	//发送的块
	sendFile           map[uint32] map[int64] bool  //发送文件块先放这里，如果发完了,则清空这里
	//发送是否完成
	sendEnd           map[uint32]  bool  //
	receive map[uint32]chan *comm.SendFileRetMsg //接受到消息的管道
}

func NewFileUpload(nc *NetWork,to int64,fp string) *FileUpload{
	n:=FileUpload{
		netclient:nc,
		timeout:to,
		sendFile:make(map[uint32] map[int64] bool ),
		sendEnd:make(map[uint32] bool ),
		receive: make(map[uint32] chan *comm.SendFileRetMsg, 10),
	}
	//注册一些方法哦
	nc.AddCallBack(comm.MID_SendFileRet,n.doSendFileRet)
	return &n
}

/**
	异步文件上传接口

callback:-1:未知异常
 */
func (n *FileUpload) Upload(lp string,rp string,checktype byte,callback func(byte)){
	//1.开启上传文件请求通道
	md5,err:=utils.GetFileMd5(lp,checktype)
	if err!=nil{
		fmt.Println("md5 error",lp,err)
		callback(100)
		return
	}
	//文件大小
	filei, err := os.Lstat(lp)
	if err!=nil{
		fmt.Println("file size eror",lp,err)
		callback(100)
		return
	}
	reqid:=utils.GetNextUint()

	//同步请求
	retb,err:=n.netclient.Request(comm.NewSendFileReqMsg(reqid,filei.Size(),md5,checktype,1,rp).GetMsg())

	if err!=nil{
		fmt.Println("request SendFileReqMsg error",lp,err)
		callback(100)
		return
	}
	reqret:=comm.NewSendFileReqRetMsgByByte(retb)
	fmt.Println("ret:",reqret.RetId)
	if reqret.RetCode==0{
		//可以上传，上传文件
		n.doUpload(lp,reqret.RetId,callback)
	}
	if reqret.RetCode==1{
		callback(100)
		return
	}
	if reqret.RetCode==2{
		callback(100)
		return
	}


	callback(100)
}


//上传文件
//支持续点上传
func (n *FileUpload) doUpload(lp string,fh uint32,callback func(byte)){
	fi, err := os.Open(lp)
	if err != nil {
		callback(100)
		return
	}
	defer fi.Close()
	n.sendFile[fh]=make(map[int64] bool)
	//开启go进行读取
	go n.goSendFileRet(fh,callback)

	//一次性传送4K
	buff := make([]byte, 4096)
	var start = int64(0)
	for {
		rn, err := fi.Read(buff)
		if err != nil && err != io.EOF {
			callback(100)
			return
		}
		if rn <= 0 {
			break
		}
		//放入发送块
		n.sendMapMutex.Lock()
		n.sendFile[fh][start]=false
		n.sendMapMutex.Unlock()
		//发送
		n.netclient.SendData(comm.NewSendFileMsg(0,fh,start,buff[:rn]).GetMsg())
		start+=int64(rn)
	}
	fmt.Println("read end")
	n.sendMapMutex.Lock()
	n.sendEnd[fh]=true
	n.sendMapMutex.Unlock()
}

//线程柱塞读取
func (n *FileUpload) goSendFileRet(fileid uint32,callback func(byte)){
	//柱塞
	select {
	case sret, ok := <-n.receive[fileid]:
		if ok{
			//成功
			if sret.RetCode==1{
				n.sendMapMutex.Lock()
				delete(n.sendFile[sret.FileId],sret.Start)
				n.sendMapMutex.Unlock()

				n.sendMapMutex.RLock()
				if n.sendEnd[sret.FileId] && len(n.sendFile[sret.FileId])==0{
					callback(1)
					n.sendMapMutex.RUnlock()
					return
				}
				n.sendMapMutex.RUnlock()
			}
		}
	case <-time.After((time.Second * 10))://10秒超时
		fmt.Println("file upload time out")
		callback(100)
		return
	}
}


//发送文件块返回哦
func (n *FileUpload) doSendFileRet(msg ziface.IMessage){
	sret :=comm.NewSendFileRetMsgByByte(msg.GetData())
	n.receive[sret.FileId]<-sret
}





