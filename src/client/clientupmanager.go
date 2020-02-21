package client

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

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

//同步某个文件到服务器
func (c *ClientUpManager) SyncFile(lp string) {
	for _, fu := range c.RemoteUpLoad {
		rp := lp[strings.Index(lp, ClientConfigObj.BasePath)+len(ClientConfigObj.BasePath):]
		fu.SyncFile(lp, rp)
	}
	fmt.Println("SyncFile..", lp)
	c.SecId++
}

//删除服务器中的某个文件
func (c *ClientUpManager) DeleteRemoteFile(lp string) {
	fmt.Println("DeleteRemoteFile..", lp)

}
