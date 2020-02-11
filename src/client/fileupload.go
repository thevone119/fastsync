package client

import (
	"comm"
	"fmt"
	"io"
	"os"
	"utils"
	"zinx/ziface"
)

//文件上传类
type FileUpload struct {
	netclient *NetWork
	timeout   int64 //超时时间，秒

	sendEnd map[uint32]bool //

}

func NewFileUpload(nc *NetWork, to int64, fp string) *FileUpload {
	n := FileUpload{
		netclient: nc,
		timeout:   to,
		sendEnd:   make(map[uint32]bool),
	}
	//注册一些方法哦
	nc.AddCallBack(comm.MID_SendFileRet, n.doSendFileRet)
	return &n
}

func (n *FileUpload) SyncFile(lp string, rp string) {

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

//发送文件块返回哦
func (n *FileUpload) doSendFileRet(msg ziface.IMessage) {
	//sret :=comm.NewSendFileRetMsgByByte(msg.GetData())

}
