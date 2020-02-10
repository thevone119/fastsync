package server

import (
	"bytes"
	"comm"
	"fmt"
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
	zlog.Debug("KeepAlive..." ,request.GetConnection().RemoteAddr())
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
	zlog.Info("Login..." ,request.GetConnection().RemoteAddr())

	//判断账号密码是否正确，
	login:=comm.NewLoginMsgByByte(request.GetData())
	//zlog.Info("Login ..." ,login.UserName,login.Pwd)

	if(login.UserName=="admin" && login.Pwd=="admin"){
		uid:=utils.GetNextUint()
		request.GetConnection().SetProperty("SESSION_ID",uid)
		request.GetConnection().SetProperty("USER_NAME",login.UserName)
		//登录成功
		request.GetConnection().SendBuffMsg(comm.NewLoginRetMsg(uid,0).GetMsg())
		zlog.Info("Login suss..." ,login.UserName,login.Pwd)
	}else{
		//登录失败
		request.GetConnection().SendBuffMsg(comm.NewLoginRetMsg(0,1).GetMsg())
	}
}





//-----------------------------------------------------------------------------------------------------------校验文件处理
type CheckFileRouter struct {
	znet.BaseRouter
}
var checkFileMutex = new(sync.Mutex)
//
func (this *CheckFileRouter) Handle(request ziface.IRequest) {
	zlog.Debug("CheckFile..." )

	checkFileMutex.Lock()
	//校验文件
	ckcekf:=comm.NewCheckFileMsgByByte(request.GetData())
	lockkey := fmt.Sprintf("CheckFile_%s", ckcekf.Filepaht)
	_,ok:=comm.TempMap.Get(lockkey)
	if ok{
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht,3).GetMsg())
		zlog.Debug("CheckFile is lock by other", ckcekf.Filepaht)
		checkFileMutex.Unlock()
		return
	}else{
		//锁1分钟
		comm.TempMap.Put(lockkey,"",60)
	}
	checkFileMutex.Unlock()


	//fp :=comm.AppendPath(comm.SyncConfigObj.BasePath,ckcekf.Filepaht)
	fp :=comm.AppendPath("/test2",ckcekf.Filepaht)
	//如果文件不存在,则上传文件
	if hasf,_:=comm.PathExists(fp);hasf==false{
		request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht,1).GetMsg())
		zlog.Info("file not found", ckcekf.Filepaht)
		return
	}else{
		//存在，则校验MD5
		md5,err:=utils.GetFileMd5(fp,ckcekf.CheckType)
		if err!=nil{
			zlog.Error(err)
			return
		}
		if bytes.Equal(md5,ckcekf.Check){
			zlog.Info("check file same", ckcekf.Filepaht)
			request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht,2).GetMsg())
		}else{
			//zlog.Debug("old md5:", fmt.Sprintf("%x",ckcekf.Check))
			//zlog.Debug("new md5:", fmt.Sprintf("%x",md5))
			zlog.Info("check file is different", ckcekf.Filepaht)
			request.GetConnection().SendBuffMsg(comm.NewCheckFileRetMsg(ckcekf.Filepaht,1).GetMsg())
		}
	}
}





//------------------------------------------------------------------------------------------------------------------request同步/异步请求处理
type RequestRouter struct {
	znet.BaseRouter
}

//request同步/异步请求处理
func (this *RequestRouter) Handle(request ziface.IRequest) {
	ckcekf:=comm.NewRequestMsgMsgByByte(request.GetData())
	//zlog.Debug("Request...",ckcekf.MsgId )
	switch ckcekf.MsgId {
	case comm.MID_SendFileReq:
		zlog.Debug("Request2...",ckcekf.MsgId )
		msg:=this.HandleSendFileReq(request,ckcekf.Data)
		request.GetConnection().SendBuffMsg(comm.NewResponseMsg(ckcekf.SecId,comm.MID_SendFileReqRet,msg.GetData()).GetMsg())
		return
	case comm.MID_Login:

		return
	case comm.MID_KeepAlive:
		request.GetConnection().SendBuffMsg(comm.NewResponseMsg(ckcekf.SecId,comm.MID_KeepAlive,comm.NewKeepAliveMsg(1).GetMsg().GetData()).GetMsg())
		return
	}



}

//请求文件校验
func (this *RequestRouter) HandleSendFileReq(request ziface.IRequest,data []byte) ziface.IMessage{
	freq:=comm.NewSendFileReqMsgByByte(data)
	zlog.Debug("HandleSendFileReq.....", freq.Filepaht)
	syncf:=SyncFileHandle.GetSyncFile(request.GetConnection().GetConnID(), freq.ReqId,freq.Filepaht,freq.Flen)
	//不是同一个客户端ID，则锁住
	if syncf.ClientId!=request.GetConnection().GetConnID(){
		zlog.Debug("SendFileReq is lock by other client", freq.Filepaht)
		return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,1).GetMsg()
	}
	//不是同一个请求，则锁住
	if syncf.ReqId!=freq.ReqId{
		zlog.Debug("SendFileReq is lock by other thread", freq.Filepaht)
		return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,1).GetMsg()
	}

	//如果文件不存在,则上传文件,同时打开文件等待接受文件数据
	if syncf.HasFile==false{
		zlog.Info("file not found", freq.Filepaht)
		syncf.Open()
		return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,0).GetMsg()
	}else{
		//存在，则校验MD5
		md5,err:=utils.GetFileMd5(syncf.FileAPath,freq.CheckType)
		if err!=nil{
			zlog.Error("check md5 err:",err)
			SyncFileHandle.RemoveSyncFile(syncf)
			return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,1).GetMsg()
		}
		if bytes.Equal(md5,freq.Check){
			zlog.Info("check file same", freq.Filepaht)
			SyncFileHandle.RemoveSyncFile(syncf)
			return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,2).GetMsg()
		}else{
			zlog.Info("check file is different", freq.Filepaht)
			syncf.Open()
			return comm.NewSendFileReqRetMsg(freq.ReqId,syncf.FileId,0).GetMsg()
		}
	}
}


//------------------------------------------------------------------------------------------------------------------SendFileMsg发送文件块的接受处理
type SendFileMsgRouter struct {
	znet.BaseRouter
}

//
func (this *SendFileMsgRouter) Handle(request ziface.IRequest) {
	zlog.Debug("SendFileMsg..." )
	sf:=comm.NewSendFileMsgByByte(request.GetData())
	syncf,ok:= SyncFileHandle.GetSyncFileById(sf.FileId)
	if ok==false{
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		zlog.Error("SendFileMsg error SyncFileHandle null", sf.FileId,syncf.FileAPath)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId,sf.FileId,sf.Start,2).GetMsg())
		return
	}
	zlog.Debug("SendFileMsg..." ,sf.Start)
	//0:写入成功  1：写入成功，并且已写入结束  2：写入失败
	wret:=syncf.Write(sf)
	switch wret{
	case 0:
		zlog.Debug("write succ..." )
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId,sf.FileId,sf.Start,1).GetMsg())
		return
	case 1:
		zlog.Debug("write succ...and write end" ,syncf.FileAPath)
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId,sf.FileId,sf.Start,1).GetMsg())
		return
	case 2:
		zlog.Error("write Error..." ,syncf.FileAPath)
		syncf.Close()
		SyncFileHandle.RemoveSyncFile(syncf)
		request.GetConnection().SendBuffMsg(comm.NewSendFileRetMsg(sf.SecId,sf.FileId,sf.Start,2).GetMsg())
		return
	}
}

