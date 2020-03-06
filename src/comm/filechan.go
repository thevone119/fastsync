package comm

import (
	"bufio"
	"container/list"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//1.实现文件的流式读取。
//每次读取文件一行，如果文件发生变化，就重新读取。读取到的记录，保存到DB中。下次直接从对应位置开始读即可。
//

//监控某个目录的所有文件，读取所有文件的行，
type FileChan struct {
	dir      string      //监控某个目录
	pattern  string      //文件名匹配规则*.fchan
	LineChan chan string //所有的文件，读取到一行，就放入到这个管道中
	isExit   bool
}

func NewFileChan(d string, p string) *FileChan {
	f := &FileChan{
		dir:      d,
		pattern:  p,
		LineChan: make(chan string, 100),
		isExit:   false,
	}
	go f.goHandle()
	return f
}

//携程轮询处理
func (f *FileChan) goHandle() {
	for {
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
			fl.open()
			l, err := fl.ReadLines()
			if err != nil {
				return nil
				//出错了,
			} else {
				for i := l.Front(); i != nil; i = i.Next() {
					f.LineChan <- i.Value.(string)
				}
			}

			return nil
		})
		if f.isExit {
			return
		}
		//1秒轮询
		time.Sleep(time.Second)
	}
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
	reader       *bufio.Reader
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
	s,err:=FileDB.GetInt64("FileLine",fp)
	if err!=nil&&s>0{
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
		f.reader = bufio.NewReader(f.FH)
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

func (f *FileLine) ReadLines() (*list.List, error) {
	f.open()
	defer f.close()
	retl := list.New()
	for {
		l, err := f.ReadLine()
		if err != nil {
			break
		}
		retl.PushBack(l)
	}
	FileDB.PutInt64("FileLine",f.fp,f.filestart)
	return retl, nil
}

//读取某行
func (f *FileLine) ReadLine() (string, error) {
	if !f.FOpen {
		return "", errors.New("file on open")
	}
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
		return "", err
	}
	f.reader.Reset(f.FH)

	l, err := f.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	sk, _ = f.FH.Seek(0, io.SeekCurrent)
	sk -= int64(f.reader.Buffered())
	f.filestart = sk

	return strings.TrimRight(l, "\n"), nil
}
