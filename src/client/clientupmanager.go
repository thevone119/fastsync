package client

import (
	"comm"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"
	"zinx/zlog"
)

//所有客户端的文件操作，均调用此类方法完成
//1.新增，修改本机文件
//2.删除本机文件
//3.

//定义全局的upload线程等待，只有等待到所有的线程结束，才推出
var SyncFileWG sync.WaitGroup

//文件上传管理类，一次上传文件到多个服务器
type ClientUpManager struct {
	RemoteUpLoad []*FileUpload //远端的服务器列表
	//发送队列
	SecId uint32 //序号ID
}

//装载所有的配置，启动所有的客户端监听，保持长连接
func NewClientUpManager() *ClientUpManager {
	cm := &ClientUpManager{
		RemoteUpLoad: make([]*FileUpload, len(ClientConfigObj.RemotePath)),
	}

	//装置客户端上传类
	for i := 0; i < len(ClientConfigObj.RemotePath); i++ {
		cm.RemoteUpLoad[i] = srtToFileUpload(ClientConfigObj.RemotePath[i])
	}
	return cm
}

//通过str构建一个fileUpload对象
//username|pwd|127.0.0.1:9001/
func srtToFileUpload(str string) *FileUpload {
	strs := strings.Split(str, "|")
	username := strs[0]
	pwd := strs[1]
	ip := strs[2][:strings.Index(strs[2], ":")]
	port, _ := strconv.Atoi(strs[2][strings.Index(strs[2], ":")+1 : strings.Index(strs[2], "/")])
	path := strs[2][strings.Index(strs[2], "/"):]
	c := NewNetWork(ip, port, username, pwd)
	fupload := NewFileUpload(c, 20, path)
	return fupload
}

//同步某个目录到服务器中去/包括目录下的所有文件
func (c *ClientUpManager) SyncPath(ltime int64, lp string, filecheck comm.CheckFileType) {

	if strings.LastIndex(lp, "/") == len(lp)-1 && len(lp) > 0 {
		lp = lp[0 : len(lp)-1]
	}
	rd, err := ioutil.ReadDir(lp)
	if err != nil {
		return
	}

	for _, fi := range rd {
		if fi.IsDir() { // 如果是目录，则回调
			fullDir := lp + "/" + fi.Name()
			c.SyncPath(ltime, fullDir, filecheck)
			continue
		} else {
			fullName := lp + "/" + fi.Name()
			c.SyncFile(ltime, fullName, filecheck)
		}
	}
}

//同步某个文件到服务器,本机文件新增，修改的时候，就调用这个方法
//cktype:文件的校验类型 //0:不校验  1:size校验 2:fastmd5  3:fullmd5
func (c *ClientUpManager) SyncFile(ltime int64, lp string, cktype comm.CheckFileType) {
	//错误拦截,针对上传过程中遇到的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("文件上传处理发生意外错误", p, lp)
		}
	}()
	zlog.Debug("SyncFile..", lp)
	//1.读取判断本地文件是否存在，大小，MD5等
	rlp, err := ClientConfigObj.GetRelativePath(lp)
	if err != nil {
		zlog.Error("SyncFile..err,path:", lp)
		return
	}
	ul := NewLocalFile(lp, rlp, cktype)
	//
	if ltime > 0 && (time.Now().Unix()-ul.FlastModTime) > ltime {
		return
	}
	for _, fu := range c.RemoteUpLoad {
		fu.SendUpload(ul)
	}
	//这里如果发完了，这个LocalFile要调用关闭方法，释放资源，释放文件的。

}

//删除服务器中的某个文件,包括文件夹
func (c *ClientUpManager) DeleteFile(lp string) {
	//不允许删除操作
	if !ClientConfigObj.AllowDel {
		zlog.Debug("AllowDel", ClientConfigObj.AllowDel, lp)
		return
	}
	zlog.Debug("DeleteFile..", lp)
	rp, err := ClientConfigObj.GetRelativePath(lp)
	if err != nil {
		zlog.Error("DeleteFile..err,path:", lp)
		return
	}
	for _, fu := range c.RemoteUpLoad {
		fu.DeleteFile(rp)
	}
}

//复制服务器的文件，包括文件夹
func (c *ClientUpManager) CopyFile(srcp string, dstp string) {
	zlog.Debug("CopyFile..", srcp, dstp)
	rsrcp, err := ClientConfigObj.GetRelativePath(srcp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", srcp)
		return
	}
	rdstp, err := ClientConfigObj.GetRelativePath(dstp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", dstp)
		return
	}
	for _, fu := range c.RemoteUpLoad {
		fu.CopyFile(rsrcp, rdstp)
	}
}

//移动服务器的文件，包括文件夹
func (c *ClientUpManager) MoveFile(srcp string, dstp string) {
	zlog.Debug("MoveFile..", srcp, dstp)
	rsrcp, err := ClientConfigObj.GetRelativePath(srcp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", srcp)
		return
	}
	rdstp, err := ClientConfigObj.GetRelativePath(dstp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", dstp)
		return
	}
	for _, fu := range c.RemoteUpLoad {
		fu.MoveFile(rsrcp, rdstp)
	}
}

func (c *ClientUpManager) Close() {
	//装置客户端上传类
	for _, fu := range c.RemoteUpLoad {
		fu.netclient.Disconnect()
	}
}
