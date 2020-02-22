package client

import (
	"comm"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"time"
	"utils"
	"zinx/ziface"
)

//文件上传类
//每个远程链接对应一个这样的类，保持长链接的文件传输处理

type FileUpload struct {
	netclient *NetWork        //网络连接类，保持长连接
	timeout   int64           //超时时间，秒
	RPath     string          //远端服务器的路径
	sendEnd   map[uint32]bool //

	upLoads chan *LocalFile //上传文件的管道 10

	//请求上传文件的管道，接受返回
	sendFileReqRetChan chan *SendFileReqRetMsg

	secId uint32
}

func NewFileUpload(nc *NetWork, to int64, fp string) *FileUpload {
	n := FileUpload{
		netclient:          nc,
		timeout:            to,
		RPath:              fp,
		sendEnd:            make(map[uint32]bool),
		upLoads:            make(chan *LocalFile, 10),
		sendFileReqRetChan: make(chan *SendFileReqRetMsg, 2),
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

//管道上传
func (n *FileUpload) doUploadChan(l *LocalFile) {
	if n.secId >= math.MaxUint32 {
		n.secId = 1
	}
	n.secId++
	_secId := n.secId

	//1.同步请求，请求服务器，看是否需要上传，如果需要上传
	n.netclient.SendData(comm.NewSendFileReqMsg(_secId, l.Flen, l.FileMd5, l.cktype, 1, l.RPath).GetMsg())

	//2.柱塞等待返回，10秒超时
	for {
		select {
		case data, ok := <-n.requestChan[_secId%10]:
			if ok {
				if data.SecId != _secId {
					break
				}
				return data.Data, nil
			}
		case <-time.After(time.Duration(10) * time.Second):
			return nil, errors.New("request time out")
		}
	}

	retb, err := n.netclient.Request(comm.NewSendFileReqMsg(_secId, l.Flen, l.FileMd5, l.cktype, 1, l.RPath).GetMsg())
}

/**
	异步文件上传接口

callback:-1:未知异常
*/
func (n *FileUpload) Upload(lp string, rp string, checktype byte, callback func(byte)) {
	//1.开启上传文件请求通道
	md5, err := utils.GetFileMd5(lp, checktype)
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
	retb, err := n.netclient.Request(comm.NewSendFileReqMsg(reqid, filei.Size(), md5, checktype, 1, rp).GetMsg())

	if err != nil {
		fmt.Println("request SendFileReqMsg error", lp, err)
		callback(100)
		return
	}
	reqret := comm.NewSendFileReqRetMsgByByte(retb)
	fmt.Println("ret:", reqret.RetId)
	if reqret.RetCode == 0 {
		//可以上传，上传文件
		n.doUpload(lp, reqret.RetId, callback)
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
func (n *FileUpload) doUpload(lp string, fh uint32, callback func(byte)) {
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
	n.sendEnd[fh] = true
	callback(1)
}

//发送文件块返回哦,返回上传完整，则记录成功，否则记录失败之类的
//这里做日志记录处理的
func (n *FileUpload) doSendFileRet(msg ziface.IMessage) {
	//sret :=comm.NewSendFileRetMsgByByte(msg.GetData())

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
