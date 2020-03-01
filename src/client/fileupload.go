package client

import (
	"comm"
	"fmt"
	"io"
	"math"
	"os"
	"time"
	"utils"
	"zinx/ziface"
	"zinx/zlog"
)

//文件上传类
//每个远程链接对应一个这样的类，保持长链接的文件传输处理
//记录上传结果，有上传完成事件

type FileUpload struct {
	netclient *NetWork //网络连接类，保持长连接
	timeout   int64    //超时时间，秒
	RPath     string   //远端服务器的路径

	upLoads chan *LocalFile //上传文件的管道 10

	//请求上传文件的管道，接受返回 2
	sendFileReqRetChan chan *comm.SendFileReqRetMsg

	//单文件上传
	sendFileState byte   //文件上传的状态记录 0:未开始 1：正在上传  2： 上传完成 3：上传错误
	sendFilePath  string //正在传送的文件名称记录

	secId uint32
}

func NewFileUpload(nc *NetWork, to int64, fp string) *FileUpload {
	n := FileUpload{
		netclient:          nc,
		timeout:            to,
		RPath:              fp,
		upLoads:            make(chan *LocalFile, 10),
		sendFileReqRetChan: make(chan *comm.SendFileReqRetMsg, 2),
		secId:              0,
	}
	//注册一些方法哦
	nc.AddCallBack(comm.MID_SendFileRet, n.doSendFileRet)
	nc.AddCallBack(comm.MID_SendFileReqRet, n.doSendFileReqRet)

	go n.goupLoadProcess()
	return &n
}

func (n *FileUpload) SyncFile(lp string, rp string) {

}

//发送上传
func (n *FileUpload) SendUpload(l *LocalFile) {
	n.upLoads <- l
}

//协程进行上传处理
func (n *FileUpload) goupLoadProcess() {
	for {
		select {
		case data, ok := <-n.upLoads:
			if ok {
				n.doUploadChan(data)
			}
		}
	}
}

//发送完成回调
func (n *FileUpload) sendEndCallBack() {

}

//管道上传，单线程
func (n *FileUpload) doUploadChan(l *LocalFile) {
	if n.secId >= math.MaxUint32 {
		n.secId = 1
	}
	n.secId++
	_secId := n.secId

	//这里要做个判断，判断客户端是否活动，如果不在活动中，这个直接就失败了。避免某个客户端连接不上，柱塞所有的任务
	if !n.netclient.IsActivity() {
		n.logUploadError(l.LPath, "服务器连接异常")
		return
	}
	//1.同步请求，请求服务器，看是否需要上传，如果需要上传
	err := n.netclient.SendData(comm.NewSendFileReqMsg(_secId, l.Flen, l.FlastModTime, l.FileMd5, l.cktype, 1, l.RPath).GetMsg())
	if err != nil {
		n.logUploadError(l.LPath, "服务器连接异常，发送数据失败")
		return
	}

	//超时时间，5秒+50M每秒（MD5校验文件，至少能达到50M/S的速度）
	timeout := 5 + l.Flen/(1024*1024*50)
	//2.柱塞等待返回，5秒超时
	for {
		select {
		case data, ok := <-n.sendFileReqRetChan:
			if ok {
				if data.ReqId != _secId {
					break
				}
				//返回了请求
				if data.RetCode == 0 {
					//可以上传，上传文件
					n.doUploadChan2(data.RetId, l)
				}
				return
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			//超时了，这里做个处理
			n.logUploadError(l.LPath, "校验文件上传超时")
			return
		}
	}
}

//管道上传2，第2步，正式上传文件，单线程
func (n *FileUpload) doUploadChan2(fh uint32, l *LocalFile) {
	//一次性传送4K
	buff := make([]byte, 4096)
	var start = int64(0)

	n.sendFileState = 1
	n.sendFilePath = l.LPath

	for {
		rn, err := l.Read(start, buff)
		if err != nil && err != io.EOF {
			n.logUploadError(l.LPath, "文件上传失败，读取文件异常")
			return
		}
		if rn <= 0 {
			break
		}
		//发送,这里直接发即可。不用缓存了，因为这个方法本来就已经有缓存
		n.netclient.SendData(comm.NewSendFileMsg(0, fh, start, buff[:rn]).GetMsg())
		start += int64(rn)
		//发送过程中，如果已经返回错误了。则直接退出哦

		if n.sendFileState == 3 {
			n.logUploadError(l.LPath, "文件上中断")
			break
		}
	}
	n.sendFileState = 2
}

/**
	异步文件上传接口

callback:-1:未知异常
*/
func (n *FileUpload) Upload_back(lp string, rp string, checktype comm.CheckFileType, callback func(byte)) {
	//1.开启上传文件请求通道
	md5, err := utils.GetFileMd5(lp, byte(checktype))
	if err != nil {
		fmt.Println("md5 error", lp, err)
		callback(100)
		return
	}
	//文件大小
	filei, err := os.Lstat(lp)
	if err != nil {
		fmt.Println("file size eror", lp, err)
		callback(100)
		return
	}
	reqid := utils.GetNextUint()

	//同步请求
	retb, err := n.netclient.Request(comm.NewSendFileReqMsg(reqid, filei.Size(), filei.ModTime().Unix(), md5, checktype, 1, rp).GetMsg())

	if err != nil {
		fmt.Println("request SendFileReqMsg error", lp, err)
		callback(100)
		return
	}
	reqret := comm.NewSendFileReqRetMsgByByte(retb)
	fmt.Println("ret:", reqret.RetId)
	if reqret.RetCode == 0 {
		//可以上传，上传文件
		n.doUpload_back(lp, reqret.RetId, callback)
	}
	if reqret.RetCode == 1 {
		callback(100)
		return
	}
	if reqret.RetCode == 2 {
		callback(100)
		return
	}

	callback(100)
}

//上传文件
//支持续点上传
func (n *FileUpload) doUpload_back(lp string, fh uint32, callback func(byte)) {
	fi, err := os.Open(lp)
	if err != nil {
		callback(100)
		return
	}
	defer fi.Close()

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
		//发送
		n.netclient.Enqueue(comm.NewSendFileMsg(0, fh, start, buff[:rn]).GetMsg())
		start += int64(rn)
	}

	callback(1)
}

//发送文件块返回哦,返回上传完整，则记录成功，否则记录失败之类的
//这里做日志记录处理的
func (n *FileUpload) doSendFileRet(msg ziface.IMessage) {
	sret := comm.NewSendFileRetMsgByByte(msg.GetData())
	if sret.RetCode == 2 {
		if n.sendFileState == 1 {
			n.sendFileState = 3
			n.logUploadError(n.sendFilePath, "文件上传错误")
		}
	}
}

//发送文件上传请求的返回
func (n *FileUpload) doSendFileReqRet(msg ziface.IMessage) {
	//sret :=comm.NewSendFileRetMsgByByte(msg.GetData())
	reqret := comm.NewSendFileReqRetMsgByByte(msg.GetData())
	n.sendFileReqRetChan <- reqret
}

//删除文件，包括文件夹
func (n *FileUpload) DeleteFile(rp string) {
	n.netclient.Enqueue(comm.NewDeleteFileRetMsg(0, 0, comm.AppendPath(n.RPath, rp)).GetMsg())
}

//复制文件，包括文件夹
func (n *FileUpload) CopyFile(srcrp string, dstrp string) {
	n.netclient.Enqueue(comm.NewMoveFileReqMsg(0, 0, comm.AppendPath(n.RPath, srcrp), comm.AppendPath(n.RPath, dstrp)).GetMsg())
}

//移动文件 ，包括文件夹
func (n *FileUpload) MoveFile(srcrp string, dstrp string) {
	n.netclient.Enqueue(comm.NewMoveFileReqMsg(0, 1, comm.AppendPath(n.RPath, srcrp), comm.AppendPath(n.RPath, dstrp)).GetMsg())
}

//记录上传成功记录
func (n *FileUpload) logUploadSuss(lp string, msg string) {

}

//记录上传失败记录
func (n *FileUpload) logUploadError(lp string, msg string) {
	zlog.Error(lp, msg)
}
