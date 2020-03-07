package client

import (
	"comm"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"
	"time"
	"utils"
)

//本地文件处理类，本地文件对应远程文件的处理
//1.实现路径的转换，本地路径适配，远程路径转换等。
//2.实现文件读取的缓存处理
//

type localFileHandle struct {
	fmap  map[string]*LocalFile //filemap
	llock sync.RWMutex          //读写锁
}

//单个最大缓存大小，1M，超过这个大小，就不缓存了，直接每次读取文件块
var MAX_CACHE_SIZE = int64(1024 * 1024)

type LocalFile struct {
	Lid          uint32             //主键
	LPath        string             //本机路径（绝对路径）
	RPath        string             //远程路径（远程的相对路径）
	cktype       comm.CheckFileType //文件校验类型
	FH           *os.File           //本机文件指针
	Flen         int64              //文件大小
	FileMd5      []byte             //文件的MD5
	CacheFile    []byte             //文件的缓存，如果文件超过1M则不进行缓存了。
	FOpen        bool               //文件是否打开
	Ferr         error              //文件读写异常
	FlastModTime int64              //文件的最后修改时间,秒
	FlastRead    int64              //文件的最后读取时间，秒,用于读超时，超过10秒没有读，就关闭文件流，避免文件卡住哦
	flock        sync.RWMutex       //读写锁

	fWait int		//正在进行上传的文件个数，如果个数等于0，则关闭文件,释放资源哦
}

func NewLocalFile(lp string, rp string, ct comm.CheckFileType) *LocalFile {
	l := &LocalFile{
		Lid:       utils.GetNextUint(),
		LPath:     lp,
		RPath:     rp,
		cktype:    ct,
		FOpen:     false,
		FH:        nil,
		Ferr:      nil,
		FileMd5:   make([]byte, 16),
		FlastRead: 0,
		fWait:0,
	}
	//这里做一些初始化等处理
	l.init()

	return l
}

//0:不校验  1:size校验 2:fastmd5  3:fullmd5
//初始化一些数据,这里考虑下异常，外层考虑吧
func (this *LocalFile) init() {
	this.FOpen = true
	//针对已存在的文件，则是打开文件，设置大小为0，并指针指向开头
	fw, err := os.Open(this.LPath)
	if err != nil {
		this.Ferr = err
		this.Close()
		return
	} else {
		this.FH = fw
	}

	fi, err := this.FH.Stat()
	if err != nil {
		this.Ferr = err
		this.Close()
		return
	}
	//文件的基础信息
	this.FlastModTime = fi.ModTime().Unix()
	this.Flen = fi.Size()

	var result []byte
	if this.Flen < MAX_CACHE_SIZE {
		//所有数据读入缓存
		this.CacheFile = make([]byte, this.Flen)
		_, err := this.FH.Read(this.CacheFile)
		if err != nil {
			this.Ferr = err
			this.Close()
			return
		}
		//缓存中对数据进行MD5
		switch this.cktype {
		case comm.FCHECK_NOT_CHECK:
		case comm.FCHECK_SIZE_CHECK:
		case comm.FCHECK_SIZE_AND_TIME_CHECK:
			//this.FileMd5 = make([]byte, 16)
			break
		case comm.FCHECK_FASTMD5_CHECK:
			hash := md5.New()
			//最多只取10块内容做MD5
			var clean = this.Flen / 10
			var start = int64(0)
			var end = int64(0)
			var temp = make([]byte, 1024)
			for {
				end = start + 1024
				if end > this.Flen {
					end = this.Flen
				}
				for i := start; i < end; i++ {
					temp[i-start] = this.CacheFile[i]
				}
				hash.Write(temp)
				start = end + clean
				if start >= this.Flen {
					break
				}
			}
			len_b := make([]byte, 8)
			binary.BigEndian.PutUint64(len_b, uint64(this.Flen))
			hash.Write(len_b)
			this.FileMd5 = hash.Sum(result)
			break
		case comm.FCHECK_FULLMD5_CHECK:
			hash := md5.New()
			hash.Write(this.CacheFile)
			this.FileMd5 = hash.Sum(result)
			break
		}
		this.Close()
	} else {
		//缓存中对数据进行MD5
		switch this.cktype {
		case comm.FCHECK_NOT_CHECK:
		case comm.FCHECK_SIZE_CHECK:
		case comm.FCHECK_SIZE_AND_TIME_CHECK:
			//this.FileMd5 = make([]byte, 16)
			break
		case comm.FCHECK_FASTMD5_CHECK:
			hash := md5.New()
			//获取文件大小
			//最多只取10块内容做MD5
			var clean = this.Flen / 10
			var temp = make([]byte, 1024)
			for {
				rn, err := this.FH.Read(temp)
				if err != nil || rn <= 0 {
					break
				}
				this.FH.Seek(clean, io.SeekCurrent)
				hash.Write(temp)
			}
			len_b := make([]byte, 8)
			binary.BigEndian.PutUint64(len_b, uint64(this.Flen))
			hash.Write(len_b)
			this.FileMd5 = hash.Sum(result)
			this.FH.Seek(0, io.SeekStart)
			break
		case comm.FCHECK_FULLMD5_CHECK:
			hash := md5.New()
			if _, err := io.Copy(hash, this.FH); err != nil {
				this.Ferr = err
				this.Close()
				return
			}
			this.FileMd5 = hash.Sum(result)
			break
		}
	}
}

func (this *LocalFile) Read(start int64, b []byte) (n int, err error) {
	this.FlastRead = time.Now().Unix()
	if this.Flen < MAX_CACHE_SIZE {
		if start >= this.Flen {
			return 0, nil
		}
		readnum := int64(len(b))
		if readnum > this.Flen-start {
			readnum = this.Flen - start
		}
		end := readnum + start
		for i := start; i < end; i++ {
			b[i-start] = this.CacheFile[i]
		}
		return int(readnum), nil
	} else {
		if !this.FOpen{
			n=0
			err= errors.New("file is close")
			return
		}
		this.flock.RLock()
		defer this.flock.RUnlock()
		this.FH.Seek(start, io.SeekStart)
		return this.FH.Read(b)
	}
}

func (this *LocalFile) AddWait(count int){
	this.flock.Lock()
	defer this.flock.Unlock()
	this.fWait+=count
}

//某个上传已经结束
func (this *LocalFile) OneUploadEnd(){
	this.flock.Lock()
	defer this.flock.Unlock()
	this.fWait--
}

//关闭文件句柄
func (this *LocalFile) Close() {
	this.FOpen = false
	if this.FH != nil {
		this.FH.Close()
		this.FH = nil
		this.CacheFile = nil
	}
}

//检测关闭方法，2秒检测一次
//如果FWait<=0,直接关闭
//如果最后读取时间超过10+FWait秒没有读取，则作为超时关闭
func (this *LocalFile) CheckAndClose() bool{
	if this.FlastRead==0{
		return false
	}
	if this.fWait<=0{
		this.Close()
		return true
	}
	if this.FlastRead+int64(this.fWait)+10 < time.Now().Unix(){
		this.Close()
		return true
	}
	return false
}
