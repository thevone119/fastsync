package comm

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zinx/zlog"
)

//1.实现文件的流式读取。
//每次读取文件一行，如果文件发生变化，就重新读取。读取到的记录，保存到DB中。下次直接从对应位置开始读即可。
//

//监控某个目录的所有文件，读取所有文件的行，
type FileChan struct {
	dir      string      //监控某个目录
	pattern  string      //文件名匹配规则*.fchan
	isExit   bool
}

func NewFileChan(d string, p string) *FileChan {
	f := &FileChan{
		dir:      d,
		pattern:  p,
		isExit:   false,
	}
	return f
}

func (f *FileChan) Start(){
	go f.goHandle()
}

//携程轮询处理
func (f *FileChan) goHandle() {
	for {
		if f.isExit {
			return
		}
		f.goHandle2()
	}
}

func (f *FileChan) goHandle2(){
	//这里避免线程轮训失败，失败后重新轮训
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("文件监控轮训发生错误，3秒后重启轮训", p,f.dir)
			time.Sleep(time.Second*3)
		}
	}()
	//通过Walk来遍历目录下f的所有子目录
	filepath.Walk(f.dir, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if f.isExit {
			return nil
		}
		m, _ := filepath.Match(f.pattern, info.Name())
		if !m {
			return nil
		}
		//获取文件
		fl := NewFileLine(p)
		if fl.FlastRedTime > info.ModTime().UnixNano() {
			return nil
		}
		fl.FlastRedTime = time.Now().UnixNano()
		err = fl.ReadLines()
		return err
	})
	//1秒轮询
	time.Sleep(time.Second)
}



//关闭所有文件，管道
func (f *FileChan) Close() {
	f.isExit = true
}

//全局变量
var _file_line_map = make(map[string]*FileLine)

type FileLine struct {
	fp           string
	FlastModTime int64    //文件的最后修改时间,秒
	FlastRedTime int64    //文件的最后读取时间,纳秒
	FOpen        bool     //文件是否打开
	filestart    int64    //当前文件位置，这个保存到数据库中的。
	FH           *os.File //本机文件指针,只读方式打开
}

func NewFileLine(fp string) *FileLine {
	fl, ok := _file_line_map[fp]
	if ok {
		return fl
	}
	fl = &FileLine{
		fp:           fp,
		FOpen:        false,
		filestart:    0,
		FlastModTime: 0,
	}
	s,err:=LeveldbDB.GetInt64([]byte("FL_"+fp))

	if err==nil&&s>0{
		fl.filestart=s
	}
	_file_line_map[fp] = fl
	return fl
}

func (f *FileLine) open() {
	if f.FOpen {
		return
	}
	fr, err := os.Open(f.fp)
	if err != nil {
		fr.Close()
		return
	} else {
		f.FH = fr
		f.FOpen = true
	}
}

func (f *FileLine) close() {
	if !f.FOpen {
		return
	}
	if f.FH != nil {
		f.FH.Close()
		f.FH = nil
		f.FOpen = false
	}
}

//处理某行记录，空格隔开
func (f *FileLine) doLine(l string) {
	s:=strings.Split(l, " ")
	for _, v := range s {
		if strings.Index(v,BASE_MON_PATH)>=0{
			FileChangeMonitorObj.AddPath(v)
		}
	}
}

//一次读取多行
func (f *FileLine) ReadLines() ( error) {
	if !f.FOpen {
		return  errors.New("file on open")
	}
	f.open()
	defer f.close()
	//
	//超过最大的，就是文件被重置了。重新来读
	if f.filestart > 0 {
		len, _ := f.FH.Seek(0, io.SeekEnd)
		if len < f.filestart {
			f.filestart = 0
		}
	}
	sk, err := f.FH.Seek(f.filestart, io.SeekStart)
	//超标
	if err != nil {
		return  err
	}
	bf:=bufio.NewReader(f.FH)

	for {
		l, err := bf.ReadString('\n')
		if err != nil {
			break
		}
		l=strings.TrimRight(l, "\n")
		if strings.Index(l,BASE_MON_PATH)>=0{
			f.doLine(l)
		}
	}
	sk, _ = f.FH.Seek(0, io.SeekCurrent)
	f.filestart = sk
	LeveldbDB.PutInt64([]byte("FL_"+f.fp),f.filestart)
	return  nil
}

