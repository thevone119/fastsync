package comm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"zinx/zlog"
)

//实现日志文件，通配文件的监控
type LogWatch struct {
	basepath   string	//监控路径，可配置目录，文件，或者文件通配
	sep string
}

func NewLogWatch(bpath string,sep string) *LogWatch {
	return &LogWatch{
		basepath: bpath,
		sep:sep,
	}
}



func (f *LogWatch) Start(){
	go f.goHandle()
}

//携程轮询处理
func (f *LogWatch) goHandle() {
	for {
		f.goHandle2()
	}
}

func (f *LogWatch) goHandle2(){
	//这里避免线程轮训失败，失败后重新轮训
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("文件监控轮训发生错误，5秒后重启轮训", p,f.basepath)
			time.Sleep(time.Second*5)
		}
	}()
	//1秒轮询
	time.Sleep(time.Second)
	//判断是文件还是目录
	fi,err:=os.Stat(f.basepath)
	//不存在，可能是通配，如/home/ap/xxxx*.log
	if err!=nil{
		fps:=filepath.Dir(f.basepath)

		fp,err:=os.Stat(fps)
		if err!=nil{
			return
		}

		pattern:=filepath.Base(f.basepath)
		if fp.IsDir(){
			files, _ := ioutil.ReadDir(fps)
			for _, onefile := range files {
				if( onefile.IsDir() ){
					continue
				}

				m, _ := filepath.Match(pattern, onefile.Name())
				if !m {
					continue
				}
				//获取文件
				fl := NewFileLine(fps+"/"+onefile.Name(),f.sep)

				//这里重新获取一次文件信息，通过路径获取的文件信息可能存在偏差
				fi,err:=os.Stat(fps+"/"+onefile.Name())
				if err==nil && fi!=nil{
					if fl.FlastRedTime > onefile.ModTime().UnixNano() && fl.FlastRedTime>fi.ModTime().UnixNano()  {
						continue
					}
				}else{
					if fl.FlastRedTime > onefile.ModTime().UnixNano()   {
						continue
					}
				}

				fl.FlastRedTime = time.Now().UnixNano()
				fl.ReadLines(1000)
			}
		}
		return
	}
	if fi.IsDir(){
		//这里不能直接配置目录，必须配置文件或者通配文件路径
	}else{
		//获取文件
		fl := NewFileLine(f.basepath,f.sep)
		if fl.FlastRedTime > fi.ModTime().UnixNano() {
			return
		}
		fl.FlastRedTime = time.Now().UnixNano()
		fl.ReadLines(1000)
	}
}



