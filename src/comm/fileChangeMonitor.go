package comm

import (
	"container/list"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//定义全局变量来进行管理，相当于静态方法
var FileChangeMonitorObj *FileChangeMonitor


/*
	提供init方法，默认加载
*/
func init() {
	FileChangeMonitorObj = &FileChangeMonitor{
		mfile:make( map[string]*FFileInfo),
		basepath:strings.Replace(BASE_PATH,"\\","/",-1),
	}
}

//文件变化监控处理类
//所有文件变化，均放入此类进行统一缓存,过滤等处理
//1.过滤重复的文件[map]过滤
//2.实现对文件重复上传的过滤
//3.对失败的文件进行重传
//4.携程循环对文件上传进行处理,1秒轮询
type FileChangeMonitor struct {
	mfile       map[string]*FFileInfo
	flock        sync.RWMutex       //读写锁
	basepath 	string
}

//添加某行记录，这行记录中记录这路径等信息
func (f *FileChangeMonitor) AddLine(l string){
	//判断路径中是否存在根路径，如果不存在，则直接过滤掉
	l=strings.Replace(l,"\\","/",-1)

	if strings.Index(l,f.basepath)<0{
		return
	}

	ps := strings.Split(l, " ")
	for _,pt := range ps {
		if strings.Index(pt, f.basepath) < 0 {
			continue
		}
		f.AddPath(pt)
	}
}


//添加文件、目录到此
//1.存在的不重复添加
//2.上传后，保留2秒，然后删除,删除前判断文件是否发生变更，如果变更了，重新上传
//3.文件最后修改时间大于1秒（判断文件已经完整了），进行文件上传
func (f *FileChangeMonitor) AddPath(p string) {
	f.flock.Lock()
	defer f.flock.Unlock()
	if len(p)<2{
		return
	}
	p=strings.Replace(p,"\\","/",-1)

	filter:=false
	//过滤的文件,文件类型
	for _,v:= range ClientConfigObj.FilterFiles{
		if fil,_:=filepath.Match(v,filepath.Ext(p));fil{
			filter=true
			break
		}
	}
	if filter{
		return
	}
	//过滤的路径，父子目录判断
	for _,v:= range ClientConfigObj.FilterPaths{
		if strings.Index(p,v)>=0{
			filter=true
			break
		}
	}
	if filter{
		return
	}

	//如果存在了，就不加入这里了，更新一个读取时间？
	mf,ok:=f.mfile[p]
	if ok{
		mf.ReadTime=time.Now().UnixNano()
		return
	}else{
		//初始化
		mf = NewFFileInfo(p)
		mf.ReadTime=time.Now().UnixNano()
		mf.ReLoadBase()
		if !mf.IsExist{
			if mf.IsDir{
				if !ClientConfigObj.AllowDelFile||!ClientConfigObj.AllowDelDir{
					return
				}
			}else{
				if !ClientConfigObj.AllowDelFile{
					return
				}
			}
			f.mfile[p]=mf
		}else{
			f.mfile[p]=mf
		}
	}
}


//获取当前队列的长度，最大长度1万，超过1万就不往里加了，等待1秒
func (f *FileChangeMonitor) GetQueueLen() int{
	f.flock.RLock()
	defer f.flock.RUnlock()
	return len(f.mfile)
}


//获取可操作的文件列表
//每秒钟获取一次，如果是修改，新增的，无变化则返回
//如果是删除的，则需要判断是否文件，目录，是否允许进行相关操作
func (f *FileChangeMonitor) GetQueue(maxl int) *list.List{
	f.flock.Lock()
	defer f.flock.Unlock()
	//800毫秒后进行处理
	ct:=time.Now().UnixNano()-1e6*1000

	l := list.New() //创建一个新的list
	//循环map
	for k, v := range f.mfile {
		if l.Len()>maxl{
			break
		}

		if v.ReadTime > ct {
			continue
		}

		//文件修改时间，大小判断
		fi,err:=os.Stat(k)
		//不存在，删除文件，文件夹操作
		if os.IsNotExist(err){
			if filepath.Ext(k) == ""{
				if !ClientConfigObj.AllowDelDir{
					delete(f.mfile, k)
					continue
				}
			}else{
				if !ClientConfigObj.AllowDelFile{
					delete(f.mfile, k)
					continue
				}
			}
			l.PushBack("del_"+k)
			delete(f.mfile, k)
		}

		if fi==nil{
			delete(f.mfile, k)
			continue
		}

		if fi.IsDir(){
			delete(f.mfile, k)
			continue
		}
		//文件，判断修改时间，大小有没有变化，没变化则推送
		if fi.ModTime().UnixNano()!=v.ModTime || fi.Size()!=v.Size{
			v.ModTime=fi.ModTime().UnixNano()
			v.Size=fi.Size()
			v.ReadTime=time.Now().UnixNano()
			continue
		}
		//符合要求的文件
		l.PushBack(k)
		delete(f.mfile, k)
	}
	return l
}





