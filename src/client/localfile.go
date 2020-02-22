package client

import (
	"os"
	"sync"
)

//本地文件处理类，本地文件对应远程文件的处理
//1.实现路径的转换，本地路径适配，远程路径转换等。
//2.实现文件读取的缓存处理，

//
type localFileHandle struct {
	fmap  map[string]*LocalFile //filemap
	llock sync.RWMutex          //读写锁
}

type LocalFile struct {
	LPath     string       //本机路径（绝对路径）
	RPath     string       //远程路径（远程的相对路径）
	cktype    byte         //文件校验类型
	FH        *os.File     //本机文件指针
	Flen      int64        //文件大小
	FileMd5   []byte       //文件的MD5
	CacheFile []byte       //文件的缓存，如果文件超过1M则不进行缓存了。
	FOpen     bool         //文件是否打开
	flock     sync.RWMutex //读写锁
}

func NewLocalFile(lp string, rp string, ct byte) *LocalFile {
	l := &LocalFile{
		LPath:  lp,
		RPath:  rp,
		cktype: ct,
		FOpen:  false,
		FH:     nil,
	}
	//这里做一些初始化等处理

	return l
}

//打开文件句柄
func (this *LocalFile) Open() error {
	if this.FOpen {
		return nil
	}
	this.FOpen = true
	//针对已存在的文件，则是打开文件，设置大小为0，并指针指向开头
	fw, err := os.Open(this.LPath)
	if err != nil {
		this.FOpen = false
		return err
	} else {
		this.FH = fw
		return nil
	}
}
