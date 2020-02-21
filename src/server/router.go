package server

import (
	"bytes"
	"comm"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
	"utils"
	"zinx/ziface"
	"zinx/zlog"
	"zinx/znet"
)

//定义服务器的所有路由处理逻辑

//KeepAliveRouter 自定义路由
type KeepAliveRouter struct {
	znet.BaseRouter
}

//Ping Handle
func (this *KeepAliveRouter) Handle(request ziface.IRequest) {
	zlog.Debug("KeepAlive...", request.GetConnection().RemoteAddr())
	//
	err := request.GetConnection().SendBuffMsg(comm.NewKeepAliveMsg(time.Now().Unix()).GetMsg())
	if err != nil {
		zlog.Error(err)
	}
}

//登录的Router 自定义路由
type LoginRouter struct {
	znet.BaseRouter
}

//Ping Handle
func (this *LoginRouter) Handle(request ziface.IRequest) {
	zlog.Info("Login...", request.GetConnection().RemoteAddr())

	//判断账号密码是否正确，
	login := comm.NewLoginMsgByByte(request.GetData())
	//zlog.Info("Login ..." ,login.UserName,login.Pwd)

	if login.UserName == "admin" && login.Pwd == "admin" {
		uid := utils.GetNextUint()
		request.GetConnection().SetProperty("SESSION_ID", uid)
		request.GetConnection().SetProperty("USER_NAME", login.UserName)
		//登录成功
		request.GetConnection().SendBuffMsg(comm.NewLoginRetMsg(uid, 0).GetMsg())
		zlog.Info("Login suss...", login.UserName, login.Pwd)
	} else {
		//登录失败
		request.GetConnection().SendBuffMsg(comm.NewLoginRetMsg(0, 1).GetMsg())
	}
}

//-----------------------------------------------------------------------------------------------------------校验文件处理
type CheckFileRouter struct {
	znet.BaseRouter
}

var checkFileMutex = new(sync.Mutex)

//
func (this *CheckFileRouter) Handle(request ziface.IRequest) {
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("CheckFile err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg("", 3).GetMsg())
		return
	}
	zlog.Debug("CheckFile...")
	checkFileMutex.Lock()
	//校验文件
	ckcekf := comm.NewCheckFileMsgByByte(request.GetData())
	lockkey := fmt.Sprintf("CheckFile_%s", ckcekf.Filepaht)
	_, ok := comm.TempMap.Get(lockkey)
	if ok {
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht, 3).GetMsg())
		zlog.Debug("CheckFile is lock by other", ckcekf.Filepaht)
		checkFileMutex.Unlock()
		return
	} else {
		//锁1分钟
		comm.TempMap.Put(lockkey, "", 60)
	}
	checkFileMutex.Unlock()

	//fp :=comm.AppendPath(comm.SyncConfigObj.BasePath,ckcekf.Filepaht)
	fp := comm.AppendPath("/test2", ckcekf.Filepaht)
	//如果文件不存在,则上传文件
	if hasf, _ := comm.PathExists(fp); hasf == false {
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht, 1).GetMsg())
		zlog.Info("file not found", ckcekf.Filepaht)
		return
	} else {
		//存在，则校验MD5
		md5, err := utils.GetFileMd5(fp, ckcekf.CheckType)
		if err != nil {
			zlog.Error(err)
			return
		}
		if bytes.Equal(md5, ckcekf.Check) {
			zlog.Info("check file same", ckcekf.Filepaht)
			request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht, 2).GetMsg())
		} else {
			//zlog.Debug("old md5:", fmt.Sprintf("%x",ckcekf.Check))
			//zlog.Debug("new md5:", fmt.Sprintf("%x",md5))
			zlog.Info("check file is different", ckcekf.Filepaht)
			request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht, 1).GetMsg())
		}
	}
}

//-----------------------------------------------------------------------------------------------------------SendFileReq,请求上传某个文件
type SendFileReqRouter struct {
	znet.BaseRouter
}

//
func (this *SendFileReqRouter) Handle(request ziface.IRequest) {
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("SendFileReq err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg("", 3).GetMsg())
		return
	}
	zlog.Debug("SendFileReq...")
	freq := comm.NewSendFileReqMsgByByte(request.GetData())

	syncf := SyncFileHandle.GetSyncFile(request.GetConnection().GetConnID(), freq.ReqId, freq.Filepaht, freq.Flen)
	//不是同一个客户端ID，则锁住
	if syncf.ClientId != request.GetConnection().GetConnID() {
		zlog.Debug("SendFileReq is lock by other client", freq.Filepaht)
		request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg())
		return
	}
	//不是同一个请求，则锁住
	if syncf.ReqId != freq.ReqId {
		zlog.Debug("SendFileReq is lock by other thread", freq.Filepaht)
		request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg())
		return
	}

	//如果文件不存在,则上传文件,同时打开文件等待接受文件数据
	if syncf.HasFile == false {
		zlog.Info("file not found", freq.Filepaht)
		syncf.Open()
		request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 0).GetMsg())
		return
	} else {
		//存在，则校验MD5
		md5, err := utils.GetFileMd5(syncf.FileAPath, freq.CheckType)
		if err != nil {
			zlog.Error("check md5 err:", err)
			SyncFileHandle.RemoveSyncFile(syncf)
			request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg())
			return
		}
		if bytes.Equal(md5, freq.Check) {
			zlog.Info("check file same", freq.Filepaht)
			SyncFileHandle.RemoveSyncFile(syncf)
			request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 2).GetMsg())
			return
		} else {
			zlog.Info("check file is different", freq.Filepaht)
			syncf.Open()
			request.GetConnection().SendBuffMsg(comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 0).GetMsg())
			return
		}
	}

}

//------------------------------------------------------------------------------------------------------------------request同步/异步请求处理
type RequestRouter struct {
	znet.BaseRouter
}

//request同步/异步请求处理
func (this *RequestRouter) Handle(request ziface.IRequest) {
	ckcekf := comm.NewRequestMsgMsgByByte(request.GetData())
	//判断是否登录了
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("Request err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(ckcekf.SecId, 1, "用户未登录", 0, "").GetMsg())
		return
	}

	//zlog.Debug("Request...",ckcekf.MsgId )
	switch ckcekf.MsgId {
	case comm.MID_SendFileReq:
		zlog.Debug("Request2...", ckcekf.MsgId)
		msg := this.HandleSendFileReq(request, ckcekf.Data)
		request.GetConnection().SendBuffMsg(comm.NewResponseMsg(ckcekf.SecId, comm.MID_SendFileReqRet, msg.GetData()).GetMsg())
		return
	case comm.MID_Login:

		return
	case comm.MID_KeepAlive:
		request.GetConnection().SendBuffMsg(comm.NewResponseMsg(ckcekf.SecId, comm.MID_KeepAlive, comm.NewKeepAliveMsg(1).GetMsg().GetData()).GetMsg())
		return
	}
}

//请求文件校验
func (this *RequestRouter) HandleSendFileReq(request ziface.IRequest, data []byte) ziface.IMessage {
	freq := comm.NewSendFileReqMsgByByte(data)
	zlog.Debug("HandleSendFileReq.....", freq.Filepaht)
	syncf := SyncFileHandle.GetSyncFile(request.GetConnection().GetConnID(), freq.ReqId, freq.Filepaht, freq.Flen)
	//不是同一个客户端ID，则锁住
	if syncf.ClientId != request.GetConnection().GetConnID() {
		zlog.Debug("SendFileReq is lock by other client", freq.Filepaht)
		return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg()
	}
	//不是同一个请求，则锁住
	if syncf.ReqId != freq.ReqId {
		zlog.Debug("SendFileReq is lock by other thread", freq.Filepaht)
		return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg()
	}

	//如果文件不存在,则上传文件,同时打开文件等待接受文件数据
	if syncf.HasFile == false {
		zlog.Info("file not found", freq.Filepaht)
		syncf.Open()
		return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 0).GetMsg()
	} else {
		//存在，则校验MD5
		md5, err := utils.GetFileMd5(syncf.FileAPath, freq.CheckType)
		if err != nil {
			zlog.Error("check md5 err:", err)
			SyncFileHandle.RemoveSyncFile(syncf)
			return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 1).GetMsg()
		}
		if bytes.Equal(md5, freq.Check) {
			zlog.Info("check file same", freq.Filepaht)
			SyncFileHandle.RemoveSyncFile(syncf)
			return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 2).GetMsg()
		} else {
			zlog.Info("check file is different", freq.Filepaht)
			syncf.Open()
			return comm.NewSendFileReqRetMsg(freq.ReqId, syncf.FileId, 0).GetMsg()
		}
	}
}

//------------------------------------------------------------------------------------------------------------------SendFileMsg发送文件块的接受处理
type SendFileMsgRouter struct {
	znet.BaseRouter
}

//
func (this *SendFileMsgRouter) Handle(request ziface.IRequest) {
	zlog.Debug("SendFileMsg...")
	sf := comm.NewSendFileMsgByByte(request.GetData())
	//判断是否登录了
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("Request err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "用户未登录", 0, "").GetMsg())
		return
	}
	syncf, ok := SyncFileHandle.GetSyncFileById(sf.FileId)
	if ok == false {
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		zlog.Error("SendFileMsg error SyncFileHandle null", sf.FileId, syncf.FileAPath)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId, sf.FileId, sf.Start, 2).GetMsg())
		return
	}
	zlog.Debug("SendFileMsg...", sf.Start)
	//0:写入成功  1：写入成功，并且已写入结束  2：写入失败
	wret := syncf.Write(sf)
	switch wret {
	case 0:
		zlog.Debug("write succ...")
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId, sf.FileId, sf.Start, 1).GetMsg())
		return
	case 1:
		zlog.Debug("write succ...and write end", syncf.FileAPath)
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId, sf.FileId, sf.Start, 1).GetMsg())
		return
	case 2:
		zlog.Error("write Error...", syncf.FileAPath)
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId, sf.FileId, sf.Start, 2).GetMsg())
		return
	}
}

//------------------------------------------------------------------------------------------------------------------DeleteFileMsg删除文件处理
type DeleteFileRouter struct {
	znet.BaseRouter
}

//
func (this *DeleteFileRouter) Handle(request ziface.IRequest) {
	zlog.Debug("DeleteFileMsg...")
	sf := comm.NewDeleteFileReqMsgByByte(request.GetData())
	//判断是否登录了
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("Request err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "用户未登录", 0, "").GetMsg())
		return
	}
	//删除文件
	err = os.Remove(sf.Filepath)

	if err != nil {
		// 删除失败
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "删除失败", 0, "").GetMsg())
	} else {
		// 删除成功
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 0, "", 0, "").GetMsg())
	}
}

//------------------------------------------------------------------------------------------------------------------MoveFileRouter移动文件处理
type MoveFileRouter struct {
	znet.BaseRouter
}

//
func (this *MoveFileRouter) Handle(request ziface.IRequest) {
	zlog.Debug("MoveFileMsg...")
	sf := comm.NewMoveFileReqMsgByByte(request.GetData())

	//判断是否登录了
	_, err := request.GetConnection().GetProperty("SESSION_ID")
	if err != nil {
		zlog.Debug("Request err no login...")
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "用户未登录", 0, "").GetMsg())
		return
	}

	//判断文件/目录是否存在
	files, err := os.Stat(sf.SrcFilepath)
	if err != nil {
		// 移动失败
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "移动失败", 0, "").GetMsg())
		return
	}

	//如果是文件夹，递归调用
	if files.IsDir() {
		err = this.copyDir(sf.SrcFilepath, sf.DstFilepath)
	} else {
		//先创建目录，再复制
		os.MkdirAll(sf.DstFilepath[:strings.LastIndex(sf.DstFilepath, "/")], os.ModePerm)
		err = this.copyFile(sf.SrcFilepath, sf.DstFilepath)
	}
	if err != nil {
		// 移动失败
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 1, "移动失败", 0, "").GetMsg())
	} else {
		// 移动成功
		request.GetConnection().SendBuffMsg(comm.NewCommRetMsg(sf.SecId, 0, "", 0, "").GetMsg())
		//如果是move 则删除源
		if sf.OpType == 1 {
			os.Remove(sf.SrcFilepath)
		}
	}
}

func (this *MoveFileRouter) copyDir(src string, dst string) error {
	rd, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	//创建目标目录，如果存在则不创建，不存在则创建
	os.MkdirAll(dst, os.ModePerm)
	for _, fi := range rd {
		if fi.IsDir() { // 如果是目录，则回调
			this.copyDir(src+"/"+fi.Name(), dst+"/"+fi.Name())
			continue
		} else {
			this.copyFile(src+"/"+fi.Name(), dst+"/"+fi.Name())
		}
	}
	return nil
}

//复制文件
func (this *MoveFileRouter) copyFile(src string, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	dstwrite, err := os.Create(src)
	if err != nil {
		return err
	}
	defer dstwrite.Close()
	_, err = io.Copy(source, dstwrite)
	if err != nil {
		return err
	}
	return nil
}
