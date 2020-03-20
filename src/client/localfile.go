package client

import (
	"bytes"
	"comm"
	"container/list"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"utils"
	"zinx/zlog"
)

//全局变量
var LocalFileHandle=&localFileHandle{
	fmap:          make(map[uint32]*LocalFile),
	Refmap:          make(map[uint32]*LocalFile),
	logDay:-1,
	IsLog:true,
}
//本地文件处理类，本地文件对应远程文件的处理
//1.实现路径的转换，本地路径适配，远程路径转换等。
//2.实现文件读取的缓存处理
//
type localFileHandle struct {
	//所有的文件上传类记录在这里，用携程去监控上传类是否上传完成，完成了则需要关闭并且释放资源
	fmap          map[uint32]*LocalFile //filemap
	//失败重发的记录在这里
	Refmap          map[uint32]*LocalFile //重发列表
	llock sync.RWMutex          //读写锁
	//统计信息
	UpLoadFileCount int64		//推送文件数
	SuccUpLoadCount int64		//成功推送服务器数
	ErrUpLoadCount int64		//推送失败数
	TimeOutCount 	int64		//超时数
	nextClearTime int64                //下次清理时间

	//日记记录系统
	logDay int	//日志的日期
	logfileOpen bool	//日志文件是否打开
	//当前日志绑定的输出文件
	logfile *os.File
	//是否记录上传日志？针对全量上传的，不记录日志哦
	IsLog bool
}






//单个最大缓存大小，100k，超过这个大小，就不缓存了，直接每次读取文件块
var MAX_CACHE_SIZE = comm.ClientConfigObj.MaxFileCache


func (s *localFileHandle) AddLocalFile(lf *LocalFile){
	s.llock.Lock()
	defer s.llock.Unlock()
	s.fmap[lf.Lid]=lf
	s.UpLoadFileCount++
}

func (s *localFileHandle) GetLocalFile(lid uint32) (*LocalFile, bool) {
	s.llock.RLock()
	defer s.llock.RUnlock()
	v, ok := s.fmap[lid]
	return v, ok
}


func (s *localFileHandle) RemoveLocalFile(lf *LocalFile) {
	s.llock.Lock()
	defer s.llock.Unlock()
	delete(s.fmap, lf.Lid)
}

//其中一个上传完成了
//统计下成功数，失败数
func  (s *localFileHandle) UpLoadEndOne(lf *LocalFile){
	s.llock.Lock()
	defer s.llock.Unlock()
	//判断是否所有都上传完成，如果都完成了。则关闭文件流，记录日志，清空缓存等
	scount:=0
	ecount:=0
	for _,rcode:=range lf.RetCodes{
		if rcode==255{
			break
		}
		if rcode==0||rcode==2{
			scount++
		}else{
			ecount++
		}
	}
	//结束了
	if scount+ecount>=len(lf.RetCodes){
		s.SuccUpLoadCount += int64(scount)
		s.ErrUpLoadCount += int64(ecount)
		zlog.Debug("Sync end",lf.LPath)
		lf.Close()
		delete(s.fmap, lf.Lid)
		//如果存在错误的，则在这里把错误记录到重发列表
		if ecount>0 && lf.ReSendCount<2 && len(s.Refmap)<100{
			lf.ReSendCount++
			lf.ReSendTime=time.Now().Unix()
			s.Refmap[lf.Lid]=lf
		}
	}
}

//获取重发的列表
func  (s *localFileHandle) GetReSend() *list.List{
	s.llock.Lock()
	defer s.llock.Unlock()
	l:=list.New()
	ct:=time.Now().Unix()-60*5
	for k, v := range s.Refmap {
		if v.ReSendTime<ct{
			l.PushBack(v)
			delete(s.Refmap,k)
		}
	}
	return l
}

//打开日志文件
//日志文件路径固定为
func (w *localFileHandle) openLog(){
	day:=time.Now().YearDay()
	//日期变了。重新开一个新的文件
	if day!=w.logDay || !w.logfileOpen{
		if w.logfileOpen{
			w.logfile.Close()
			w.logfileOpen = false
		}
		os.MkdirAll(comm.TRANSFER_PATH,0755)
		//先创建目录
		logpath:=filepath.Join(comm.TRANSFER_PATH,"tlog_"+time.Now().Format("20060102_150405")+".fstlog")
		exist := true
		if _, err := os.Stat(logpath); os.IsNotExist(err) {
			exist = false
		}

		if exist {
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

//记录日志文件
func  (s *localFileHandle) Log(str string){
	if !s.IsLog{
		return
	}
	s.openLog()
	if s.logfileOpen{
		fmt.Fprintln(s.logfile,str)
	}
}

//定时清理，每10秒清理一次哦，超过10秒没有读取的文件，则关闭文件
func (s *localFileHandle) ClearTimeout() {
	ct := time.Now().Unix()
	if ct > s.nextClearTime {
		s.nextClearTime = ct + 10
		//清理
		s.llock.Lock()
		defer s.llock.Unlock()
		clearTime := ct - 10
		for k, v := range s.fmap {
			isstart:=true
			if v.FlastRead<=0{
				isstart=false
				break
			}else{
				for stime:=range v.SUpLoadTime{
					if stime<=0{
						isstart=false
						break
					}
				}
			}
			if isstart && v.FlastRead < clearTime {
				v.Close()
				delete(s.fmap, k)
			}
		}
	}
}


//各个服务器上传完成，要进行相关的汇总
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
	RetCodes 	[]byte				//各个服务器端上传返回的码，默认初始化为255
	SUpLoadTime []int64				//各个服务器端开始上传时间,毫秒
	EUpLoadTime []int64				//各个服务器端结束上传时间,毫秒

	ReSendCount int8				//重发次数
	ReSendTime int64				//加入重发里的时间
}

func NewLocalFile(slen int,lp string, rp string, ct comm.CheckFileType) *LocalFile {
	l := LocalFile{
		Lid:       utils.GetNextUint(),
		LPath:     lp,
		RPath:     rp,
		cktype:    ct,
		FOpen:     false,
		FH:        nil,
		Ferr:      nil,
		FileMd5:   make([]byte, 16),
		RetCodes:  make([]byte, slen),
		SUpLoadTime:make([]int64, slen),
		EUpLoadTime:make([]int64, slen),
		FlastRead: 0,
		ReSendCount:0,
		ReSendTime:0,

	}
	//默认都是255，还未开始上传
	for i:=0;i<slen;i++{
		l.RetCodes[i]=255
	}

	//这里做一些初始化等处理
	l.Init()

	return &l
}

func (this *LocalFile) Init(){
	//错误拦截,针对上传过程中遇到的错误进行拦截，避免出现意外错误，程序退出
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("文件信息初始化发生错误", this.LPath,p)
		}
	}()

	this.init2()
}


//0:不校验  1:size校验 2:fastmd5  3:fullmd5
//初始化一些数据,这里考虑下异常，外层考虑吧
func (this *LocalFile) init2() {
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
		this.FH.Close()
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
	if !this.FOpen{
		n=0
		err= errors.New("file is close")
		return
	}
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
		this.flock.RLock()
		defer this.flock.RUnlock()
		this.FH.Seek(start, io.SeekStart)
		return this.FH.Read(b)
	}
}


//关闭文件句柄,只关闭一次哦
func (this *LocalFile) Close() {
	if !this.FOpen{
		return
	}
	this.FOpen = false
	this.CacheFile = nil
	if this.FH != nil {
		this.FH.Close()
		this.FH = nil
	}
	//只在关闭的时候调用一次记录日志。后续根据日志实现重跑机制
	this.Log()
}

//记录上传日志，所有传输都完成了，才进行日志记录
//超时被清理也记录是上传日志
//后续实现日志数据入库机制，暂时考虑支持oracle数据库
//日志记录通过|符号分隔,服务器返回通过$符号分隔
//时间|操作|本机路径|服务器1$开始时间$结束时间$返回结果|服务器2$开始时间$结束时间$返回结果
func (this *LocalFile) Log(){
	str := bytes.Buffer{}
	str.WriteString(time.Now().Format("20060102_150405"))
	str.WriteString("|")
	str.WriteString("u")
	str.WriteString("|")
	str.WriteString(this.RPath)
	//循环输出服务器传输结果
	for i:=0;i<len(this.RetCodes);i++{
		str.WriteString("|")
		str.WriteString(comm.ClientConfigObj.GetRemoteName(i))
		str.WriteString("$")
		str.WriteString(strconv.FormatInt(this.SUpLoadTime[i],10))
		str.WriteString("$")
		str.WriteString(strconv.FormatInt(this.EUpLoadTime[i],10))
		str.WriteString("$")
		str.WriteString(strconv.FormatInt(int64(this.RetCodes[i]),10))
	}
	LocalFileHandle.Log(str.String())
}
