package comm

import (
	"sync"
)

//定义全局变量来进行管理，相当于静态方法
var FileChangeMonitorObj *FileChangeMonitor


/*
	提供init方法，默认加载
*/
func init() {
	FileChangeMonitorObj = &FileChangeMonitor{
		mfile:make( map[string]*FFileInfo),
	}
}

//文件变化监控处理类
//所有文件变化，均放入此类进行统一缓存
//1.过滤重复的文件[map]过滤
//2.实现对文件重复上传的过滤
//3.对失败的文件进行重传
//4.携程循环对文件上传进行处理,1秒轮询
type FileChangeMonitor struct {
	mfile       map[string]*FFileInfo
	flock        sync.RWMutex       //读写锁
}

//添加文件、目录到此
//1.存在的不重复添加
//2.上传后，保留2秒，然后删除,删除前判断文件是否发生变更，如果变更了，重新上传
//3.文件最后修改时间大于1秒（判断文件已经完整了），进行文件上传
func (f *FileChangeMonitor) AddPath(p string) {
	f.flock.Lock()
	defer f.flock.Unlock()

	mf,ok:=f.mfile[p]
	if ok{
		//如果存在了，就不加入这里了，更新一个读取时间？


		return
	}else{
		//初始化
		mf= NewFFileInfo(p)
		mf.ReLoadBase()

		f.mfile[p]=mf
	}
}





