package client

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"zinx/zlog"
)

//所有客户端的文件操作，均调用此类方法完成
//1.新增，修改本机文件
//2.删除本机文件
//3.

//文件上传管理类，一次上传文件到多个服务器
type ClientUpManager struct {
	RemoteUpLoad []*FileUpload //远端的服务器列表

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
func (c *ClientUpManager) SyncPath(lp string) {
	rd, err := ioutil.ReadDir(lp)
	if err != nil {
		return
	}
	for _, fi := range rd {
		if fi.IsDir() { // 如果是目录，则回调
			fullDir := lp + "/" + fi.Name()
			c.SyncPath(fullDir)
			continue
		} else {
			fullName := lp + "/" + fi.Name()
			c.SyncFile(fullName)
		}
	}
}

//同步某个文件到服务器,本机文件新增，修改的时候，就调用这个方法
func (c *ClientUpManager) SyncFile(lp string) {

	for _, fu := range c.RemoteUpLoad {
		rp := lp[strings.Index(lp, ClientConfigObj.BasePath)+len(ClientConfigObj.BasePath):]
		fu.SyncFile(lp, rp)
	}
	fmt.Println("SyncFile..", lp)
	c.SecId++
}

//删除服务器中的某个文件,包括文件夹
func (c *ClientUpManager) DeleteFile(lp string) {
	zlog.Debug("DeleteFile..", lp)
	rp, err := ClientConfigObj.GetRelativePath(lp)
	if err != nil {
		zlog.Error("DeleteFile..err,path:", lp)
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
	}
	rdstp, err := ClientConfigObj.GetRelativePath(dstp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", dstp)
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
	}
	rdstp, err := ClientConfigObj.GetRelativePath(dstp)
	if err != nil {
		zlog.Error("CopyFile..err,path:", dstp)
	}
	for _, fu := range c.RemoteUpLoad {
		fu.MoveFile(rsrcp, rdstp)
	}
}
