package client

import (
	"comm"
	"container/list"
	"os"
	"path/filepath"
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
	//所有的文件上传类记录在这里，用携程去监控上传类是否上传完成，完成了则需要关闭并且释放资源
	lfileList *list.List
}

//装载所有的配置，启动所有的客户端监听，保持长连接
func NewClientUpManager() *ClientUpManager {
	cm := &ClientUpManager{
		RemoteUpLoad: make([]*FileUpload, len(ClientConfigObj.RemotePath)),
		lfileList:list.New(),
	}

	//装置客户端上传类
	for i := 0; i < len(ClientConfigObj.RemotePath); i++ {
		cm.RemoteUpLoad[i] = srtToFileUpload(ClientConfigObj.RemotePath[i])
	}
	//开携程，处理清理工作
	go cm.goCloseLocalFile()
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
	c := NewNetWork(ip, port, username, pwd,strs[2])
	fupload := NewFileUpload(c, 20, path)
	return fupload
}

//同步某个目录到服务器中去/包括目录下的所有文件
//ltime: 秒哦
func (c *ClientUpManager) SyncPath(ltime int64, lp string, filecheck comm.CheckFileType) {
	//Walk函数会遍历root指定的目录下的文件树，对每一个该文件树中的目录和文件都会调用walkFn，包括root自身。
	//所有访问文件/目录时遇到的错误都会传递给walkFn过滤。文件是按词法顺序遍历的，这让输出更漂亮，但也导致处理非常大的目录时效率会降低。
	//Walk函数不会遍历文件树中的符号链接（快捷方式）文件包含的路径。
	currtime := time.Now().Unix()
	filepath.Walk(lp, func(path string, info os.FileInfo, err error) error {
		//这里判断是否为目录，只需监控目录即可
		//目录下的文件也在监控范围内，不需要我们一个一个加
		if info.IsDir() {
			return nil
		}
		if ltime > 0 && (currtime-info.ModTime().Unix()) > ltime {
			return nil
		}
		c.SyncFile(path, filecheck)
		return nil
	})
}

//同步某个文件到服务器,本机文件新增，修改的时候，就调用这个方法
//cktype:文件的校验类型 //0:不校验  1:size校验 2:fastmd5  3:fullmd5
func (c *ClientUpManager) SyncFile(lp string, cktype comm.CheckFileType) {
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
	//加入等待，全局等待
	SyncFileWG.Add(len(c.RemoteUpLoad))
	ul.AddWait(len(c.RemoteUpLoad))
	//放入这个监控队列
	c.lfileList.PushBack(ul)
	for _, fu := range c.RemoteUpLoad {
		fu.SendUpload(ul)
	}
	//这里如果发完了，这个LocalFile要调用关闭方法，释放资源，释放文件的。

}

//删除服务器中的某个文件,包括文件夹
func (c *ClientUpManager) DeleteFile(lp string) {
	//不允许删除操作
	if !ClientConfigObj.AllowDelFile {
		zlog.Debug("AllowDelFile", ClientConfigObj.AllowDelFile, lp)
		return
	}
	//如果是目录，则判断目录是否可以删除
	if filepath.Ext(lp) == "" && !ClientConfigObj.AllowDelDir{
		zlog.Debug("AllowDelDir", ClientConfigObj.AllowDelDir, lp)
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

//采用携程关闭本地文件，释放资源
func (c *ClientUpManager) goCloseLocalFile(){
	//1秒轮训
	for{
		if c.lfileList.Len()>0{
			var next *list.Element
			for e := c.lfileList.Front(); e != nil; e = next {
				next = e.Next()
				if e.Value.(*LocalFile).CheckAndClose(){
					c.lfileList.Remove(e)
				}
			}
		}
		time.Sleep(time.Second*1)
	}
}
