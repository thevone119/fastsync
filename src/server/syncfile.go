package server

import (
	"comm"
	"io"
	"os"
	"sync"
	"time"
	"utils"
)

//全局对象
var SyncFileHandle = &syncFileHandle{
	fmap:          make(map[string]*SyncFile),
	fhmap:         make(map[uint32]*SyncFile),
	nextClearTime: 0,
}

//实现文件写入锁，同一个时间，一个文件只允许一个客户端进行写入操作
type syncFileHandle struct {
	fmap          map[string]*SyncFile //filemap
	fhmap         map[uint32]*SyncFile //filemap
	nextClearTime int64                //下次清理时间
	flock         sync.RWMutex         //读写锁
}

func (s *syncFileHandle) GetSyncFile(cid uint32, reqid uint32, fp string, flen int64) *SyncFile {
	s.flock.Lock()
	defer s.flock.Unlock()
	v, ok := s.fmap[fp]
	if ok {
		return v
	} else {
		f := NewSyncFile(cid, reqid, fp, flen)
		s.fmap[fp] = f
		s.fhmap[f.FileId] = f
		return f
	}
}

func (s *syncFileHandle) GetSyncFileById(fileid uint32) (*SyncFile, bool) {
	s.flock.RLock()
	defer s.flock.RUnlock()
	v, ok := s.fhmap[fileid]
	return v, ok
}

func (s *syncFileHandle) RemoveSyncFile(sf *SyncFile) {
	s.flock.Lock()
	defer s.flock.Unlock()
	delete(s.fmap, sf.FilePt)
	delete(s.fhmap, sf.FileId)
}

//当客户端关闭后，情况某个客户端的捆绑的所有数据
func (s *syncFileHandle) CloseAll(cid uint32) {
	s.flock.Lock()
	defer s.flock.Unlock()
	for k, v := range s.fmap {
		if v.ClientId != cid {
			continue
		}
		v.Close()
		delete(s.fhmap, v.FileId)
		delete(s.fmap, k)
	}
}

//定时清理，每5分钟清理一次哦，超过5分钟没有操作的文件，则关闭文件
func (s *syncFileHandle) ClearTimeout() {
	ct := time.Now().Unix()
	if ct > s.nextClearTime {
		s.nextClearTime = ct + 60*5
		//清理
		s.flock.Lock()
		defer s.flock.Unlock()
		clearTime := ct - 60*5
		for k, v := range s.fmap {
			if v.LastTime < clearTime {
				v.Close()
				delete(s.fhmap, v.FileId)
				delete(s.fmap, k)
			}
		}
	}
}

//同步文件处理
type SyncFile struct {
	ClientId uint32   //客户端ID，不同客户端，不运行同占一个文件
	ReqId    uint32   //就算是同一个客户端，不同的请求ID，也不能占用一个文件
	FilePt   string   //文件路径，相对路径
	Flen     int64    //文件大小
	FileId   uint32   //文件句柄ID
	FH       *os.File //文件指针
	//
	FileAPath string //文件的绝对路径
	FOpen     bool   //文件是否已经打开
	HasFile   bool   //文件是否存在，不存在则需要新增
	//
	LastTime int64 //最后修改时间

	FlastModTime int64 //文件的最后修改时间
	//
	WriteLen int64        //已写入文件的长度，当写入文件长度等于文件长度时，写入完整
	flock    sync.RWMutex //读写锁

}

func NewSyncFile(cid uint32, reqid uint32, fp string, flen int64) *SyncFile {
	f := SyncFile{
		ClientId:  cid,
		ReqId:     reqid,
		FilePt:    fp,
		Flen:      flen,
		FileId:    utils.GetNextUint(),
		FileAPath: comm.AppendPath(ServerConfigObj.BasePath, fp),
		FH:        nil,
		FOpen:     false,
		HasFile:   false,
		LastTime:  time.Now().Unix(),
	}

	//判断文件是否存在
	if hasf, _ := comm.PathExists(f.FileAPath); hasf == false {
		f.HasFile = false
	} else {
		f.HasFile = true
	}

	return &f
}

//打开文件句柄
func (this *SyncFile) Open() error {
	if this.FOpen {
		return nil
	}
	this.FOpen = true
	//针对已存在的文件，则是打开文件，设置大小为0，并指针指向开头
	//不存在的文件，则创建文件
	fw, err := os.Create(this.FileAPath)
	if err != nil {
		this.FOpen = false
		return err
	} else {
		this.FH = fw
		return nil
	}
}

//写文件数据
//return 0:写入成功  1：写入成功，并且已写入结束  2：写入失败
func (this *SyncFile) Write(sf *comm.SendFileMsg) byte {
	//写锁
	this.flock.Lock()
	defer this.flock.Unlock()
	this.LastTime = time.Now().Unix()
	if this.FOpen == false || this.FH == nil {
		return 2
	}

	_, err := this.FH.Seek(sf.Start, io.SeekStart)
	if err != nil {
		return 2
	}
	rn, err := this.FH.Write(sf.Fbyte)
	if err != nil {
		return 2
	}
	this.WriteLen += int64(rn)
	if this.WriteLen >= this.Flen {
		return 1
	}

	return 0
}

//关闭文件句柄
func (this *SyncFile) Close() {
	this.FOpen = false
	if this.FH != nil {
		this.FH.Close()
		this.FH = nil
	}
}
