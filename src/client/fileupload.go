package client

import (
	"comm"
	"errors"
	"io"
	"math"
	"time"
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
	//发送文件块的返回（目前只接受完成的返回，中间返回不放入这里）,所以定义2大小的管道即可
	sendFileRetChan chan *comm.SendFileRetMsg

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
		sendFileReqRetChan: make(chan *comm.SendFileReqRetMsg, 5),
		sendFileRetChan:    make(chan *comm.SendFileRetMsg, 5),
		secId:              0,
	}
	//注册一些方法哦
	nc.AddCallBack(comm.MID_SendFileRet, n.doSendFileRet)
	nc.AddCallBack(comm.MID_SendFileReqRet, n.doSendFileReqRet)

	//这里要不要开2个协程提升并发?目前不能开，因为2个协程，会有返回值冲突sendFileReqRetChan冲突
	go n.goupLoadProcess()
	return &n
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
				ret, err := n.doUploadChan(data)
				//这里对发送完成做处理
				if ret == 0 {

				} else {
					zlog.Error(err)
				}
			}
		}
	}
}

//发送完成回调
func (n *FileUpload) sendEndCallBack() {

}

//异常包裹，出现任何的异常，均返回未知异常
// 0:上传成功 1：io失败，无法上传，2：文件一致，无需上传 3：服务器读写错误 4：服务器连接异常，5：服务器连接异常，发送数据失败，6：校验文件上传超时，7：文件上传失败，读取文件异常
//8：文件上中断，9：文件发送超时，10：文件块发送超时，11：服务端文件块写入失败 100：未知异常
func (n *FileUpload) doUploadChan(l *LocalFile) (retb byte, err error) {
	//错误拦截,针对上传过程中遇到的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			str, ok := p.(string)
			if ok {
				err = errors.New(str)
			} else {
				err = errors.New("panic")
			}
			retb = 100
			zlog.Error("文件上传发送意外错误", err)
		}
	}()
	return n.doUploadChan2(l)
}

//管道上传，单线程,校验是否需要上传，如果需要上传再进行第3步的文件上传
func (n *FileUpload) doUploadChan2(l *LocalFile) (byte, error) {
	if n.secId >= math.MaxUint32 {
		n.secId = 1
	}
	n.secId++
	_secId := n.secId

	//这里要做个判断，判断客户端是否活动，如果不在活动中，这个直接就失败了。避免某个客户端连接不上，柱塞所有的任务
	if !n.netclient.IsActivity() {
		return 4, errors.New("服务器连接异常")
	}

	//每次发送前先清下管道,有效避免管道柱塞
	if len(n.sendFileReqRetChan) > 0 {
		<-n.sendFileReqRetChan
	}

	//1.同步请求，请求服务器，看是否需要上传，如果需要则上传
	err := n.netclient.SendData(comm.NewSendFileReqMsg(_secId, l.Flen, l.FlastModTime, l.FileMd5, l.cktype, 1, l.RPath).GetMsg())
	if err != nil {
		return 5, errors.New("服务器连接异常，发送数据失败")
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
					return n.doUploadChan3(data.RetId, l)
				}
				if data.RetCode == 1 {
					return 1, errors.New("服务端io失败，无法上传")
				}
				if data.RetCode == 2 {
					return 1, errors.New("文件一致，无需上传")
				}
				return 100, errors.New("发送请求，未知返回码")
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			//超时了，这里做个处理
			return 6, errors.New("校验文件上传超时")
		}
	}
	return 100, errors.New("不可达错误")
}

//管道上传3，第3步，正式上传文件，单线程柱塞
func (n *FileUpload) doUploadChan3(fh uint32, l *LocalFile) (byte, error) {
	//一次性传送4K
	buff := make([]byte, 4096)
	var start = int64(0)
	n.sendFileState = 1
	n.sendFilePath = l.LPath
	//每次发送，都先清一下管道，避免特殊情况下管道的柱塞
	if len(n.sendFileRetChan) > 0 {
		<-n.sendFileRetChan
	}
	for {
		rn, err := l.Read(start, buff)
		if err != nil && err != io.EOF {
			return 7, errors.New("文件上传失败，读取发送文件异常")
		}
		if rn <= 0 {
			break
		}
		//发送,这里直接发即可。不用缓存了，因为这个方法本来就已经有缓存
		err = n.netclient.SendData(comm.NewSendFileMsg(0, fh, start, buff[:rn]).GetMsg())
		if err != nil {
			return 5, errors.New("服务器连接异常，发送数据失败,发送文件中断")
		}
		start += int64(rn)
		//发送过程中，如果已经返回错误了。则直接退出哦
		if n.sendFileState == 3 {
			return 8, errors.New("服务器端写文件失败，发送文件中断")
		}
	}
	n.sendFileState = 2
	//
	timeout := 5 + l.Flen/(1024*1024*50)
	//这里柱塞等待文件发送完成，然后返回，做超时处理。因为是文件块，因此5秒超时，这里要测试，发送完后，最后一块多久才返回，如果有发送缓存的话，怎么办呀
	//2.柱塞等待返回，5秒超时
	for {
		select {
		case data, ok := <-n.sendFileRetChan:
			if ok {
				if data.FileId != fh {
					break
				}
				//传输完成,文件上传成功了
				if data.RetCode == 3 {
					return 0, nil
				}
				if data.RetCode == 2 {
					return 11, errors.New("服务端文件块写入失败")
				}
				return 100, errors.New("未知错误，不可到的达传送文件块返回")
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			//超时了，这里做个处理
			return 10, errors.New("文件块发送超时")
		}
	}
	return 100, errors.New("不可达错误")
}

//发送文件块返回哦,返回上传完整，则记录成功，否则记录失败之类的
//这里做日志记录处理的
func (n *FileUpload) doSendFileRet(msg ziface.IMessage) {
	sret := comm.NewSendFileRetMsgByByte(msg.GetData())
	if sret.RetCode == 2 {
		if n.sendFileState == 1 {
			n.sendFileState = 3
		}
		n.sendFileRetChan <- sret
	}
	//发送完成了
	if sret.RetCode == 3 {
		n.sendFileRetChan <- sret
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
