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



var checkFileMutex = new(sync.Mutex)

//校验文件处理
type CheckFileRouter struct {
	znet.BaseRouter
}

//Ping Handle
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





//request同步/异步请求处理
type RequestRouter struct {
	znet.BaseRouter
}

//request同步/异步请求处理
func (this *RequestRouter) Handle(request ziface.IRequest) {
	zlog.Debug("Request..." )
	ckcekf:=comm.NewRequestMsgMsgByByte(request.GetData())
	switch ckcekf.MsgId {
	case comm.MID_SendFileReq:
		zlog.Debug("Request...",ckcekf.MsgId )
		msg:=this.HandleSendFileReq(request,ckcekf.Data)
		request.GetConnection().SendBuffMsg(comm.NewResponseMsg(ckcekf.SecId,comm.MID_SendFileReqRet,msg.GetData()).GetMsg())

	}



}



//请求文件校验
func (this *RequestRouter) HandleSendFileReq(request ziface.IRequest,data []byte) ziface.IMessage{
	freq:=comm.NewSendFileReqMsgByByte(data)
	//对文件校验进行处理




	return comm.NewSendFileReqRetMsg(freq.ReqId,100,0).GetMsg()

}


