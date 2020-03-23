package comm

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"time"
	"zinx/zlog"
)

//监听文件变化，这个要测试性能，看10W，100W个目录消耗内存情况。
//目前GO有2个监听包https://github.com/howeyc/fsnotify  https://github.com/fsnotify/fsnotify
//测试后，还是google的包好用点
//需要做一些完整测试。比如移动目录，好像无法监听？
//windows在监控的情况下，是无法直接删除父目录的。
//监控日志统一存放在notifylog目录,名字命名命名为xxx.nlog
type FSWatch struct {
	watch      *fsnotify.Watcher
	basepath   string
	exitChan   chan bool //退出信号
	errorCount int       //错误次数
	//当前日志绑定的输出文件
	logfile *os.File
	logfileOpen bool
	logDay int	//日志的日期
}

func NewFSWatch(bpath string) *FSWatch {
	return &FSWatch{
		watch:    nil,
		basepath: bpath,
		exitChan: make(chan bool, 2),
		logDay:-1,
	}
}

//通过字符串判断是否目录，就是最后一个如果有文件后缀就判断是目录
func (w *FSWatch) IsDir(dir string) bool {
	return filepath.Ext(dir) == ""
}
//打开日志文件
func (w *FSWatch) openLog(){
	day:=time.Now().YearDay()
	//日期变了。重新开一个新的文件
	if day!=w.logDay{
		if w.logfileOpen{
			w.logfile.Close()
			w.logfileOpen = false
		}
		os.MkdirAll(NOTIFY_PATH,0755)
		//先创建目录
		logpath:=filepath.Join(NOTIFY_PATH,"nlog_"+time.Now().Format("20060102_150405")+".fstlog")
		if checkFileExist(logpath) {
			//文件存在，打开
			f, err := os.OpenFile(logpath, os.O_APPEND|os.O_RDWR, 0644)
			if err!=nil{
				zlog.Error("open file err",logpath,err)
			}else{
				w.logfile=f
				w.logfileOpen = true
			}
		} else {
			//文件不存在，创建
			f, err := os.OpenFile(logpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err!=nil{
				zlog.Error("open file err",logpath,err)
			}else{
				w.logfile=f
				w.logfileOpen = true
			}
		}
		w.logDay=day
	}
}


func (w *FSWatch) Start() {
	fi,err:=os.Stat(w.basepath)

	if err!=nil || fi==nil ||!fi.IsDir(){
		zlog.Error("监控服务开启失败", err)
		return
	}
	wer, err := fsnotify.NewWatcher()
	if err != nil {
		zlog.Error("监控服务开启失败", err)
		return
	}
	w.watch = wer

	//开启线程对监控返回数据进行处理,先开启处理，避免添加目录的过程太长，中途数据堆积
	go w.goHandleEvent()

	count := 0
	now := time.Now()
	//通过Walk来遍历目录下的所有子目录
	filepath.Walk(w.basepath, func(path string, info os.FileInfo, err error) error {
		//这里判断是否为目录，只需监控目录即可
		//目录下的文件也在监控范围内，不需要我们一个一个加
		if info.IsDir() {
			count++
			err = w.watch.Add(path)
			if err != nil {
				return err
			}
			if count%1000==0{
				zlog.Info("已添加监控目录", count,"个")
			}
		}
		return nil
	})
	zlog.Info("监控目录服务开启", "监控根目录为:", w.basepath, "子目录数:", count, "监控服务启动耗时:", time.Now().Sub(now))

}

func (w *FSWatch) goHandleEvent() {
	//错误拦截,针对监控过程中可能出现的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			w.errorCount++
			zlog.Error("文件监控过程中发生意外错误", p, ",错误次数", w.errorCount, "监控程序将在", w.errorCount*2, "秒后重启")
			w.watch.Close()
			time.Sleep(time.Duration(w.errorCount*2) * time.Second)
			w.Start()
		}
	}()

	for {
		select {
		case ev := <-w.watch.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					zlog.Debug("创建文件：", ev.Name)
					w.Log(ev.Name, "c")
					//这里获取新创建文件的信息，如果是目录，则加入监控中,
					//1.这里还要考虑子目录的问题
					if w.IsDir(ev.Name) {
						w.watch.Add(ev.Name)
						zlog.Debug("添加监控 : ", ev.Name)
					}
				}
				if ev.Op&fsnotify.Write == fsnotify.Write {
					w.Log(ev.Name, "w")
					zlog.Debug("写入文件 : ", ev.Name)
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					zlog.Debug("删除文件 : ", ev.Name)
					w.Log(ev.Name, "d")
					//如果删除文件是目录，则移除监控
					if w.IsDir(ev.Name) {
						w.watch.Remove(ev.Name)
						zlog.Debug("删除监控 : ", ev.Name)
					}
				}
				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					w.Log(ev.Name, "r")
					zlog.Debug("重命名文件 : ", ev.Name)
					if w.IsDir(ev.Name) {
						w.watch.Remove(ev.Name)
						zlog.Debug("删除监控 : ", ev.Name)
					}
				}
				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					//zlog.Debug("修改权限 : ", ev.Name)
					//fmt.Println("修改权限 : ", ev.Name)
				}
			}
		case <-w.watch.Errors:
			{
				//return;
				//忽略错误
			}
		case <-w.exitChan:
			{
				//退出携程
				return
			}
		}
	}
}

//可以每天重启一次监控，避免特殊情况下出现某些目录没有受到监控的情况
func (w *FSWatch) Close() {
	//
	w.watch.Close()
	//让携程退出
	w.exitChan <- true
	zlog.Info("监控目录关闭")
}

//记录日志哦，通过日志进行监控中转。
//后续可使用外部其他的工具生成日志，从而生成监控
//同时也避免了事件阻塞
//日志格式标准为 2020-02-02 03:03:03 /home/ap/cc/xx.txt d
//考虑兼容inotify-tools工具生成的日志格式
func (w *FSWatch) Log(p string, op string) {
	FileChangeMonitorObj.AddPath(p)
	//不加日志了直接加到内存中
	//w.openLog()
	if w.logfileOpen{
		//fmt.Fprintln(w.logfile,time.Now().Format("2006-01-02 15:04:05"),p,op)
	}
}
