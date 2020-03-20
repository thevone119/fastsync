package comm

import (
	"bytes"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"path/filepath"
	"time"
	"zinx/zlog"
)

//实现轮训算法的文件监控。
//只监控目录下的文件，不监控目录本身
type PollWatch struct {
	basepath   []string	//轮训的目录列表
	prePollCount int64	//上一个批次
	pollCount int64		//轮训的次数，批次，通过批次，对比是否有删除操作
	tempValue int64		//临时值，避免重复申请int64
}




func NewPollWatch(bpath []string) *PollWatch {
	return &PollWatch{
		basepath: bpath,
		prePollCount:0,
		pollCount:0,
		tempValue:0,
	}
}

func (f *PollWatch) Start(){
	go f.goHandle()
}

//携程轮询处理
func (f *PollWatch) goHandle() {
	for {

		f.goHandle2()
		f.pollCount=time.Now().UnixNano()
		//5秒轮询
		time.Sleep(time.Duration(ClientConfigObj.PollTime)*time.Millisecond)
	}
}


func (f *PollWatch) goHandle2(){
	//这里避免线程轮训失败，失败后重新轮训
	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			zlog.Error("文件监控轮循发生错误，10秒后重启轮循", p,f.basepath)
			time.Sleep(time.Second*10)
		}
	}()

	if !LeveldbDB.isopen{
		zlog.Error("文件记录数据库打开失败，无法轮循目录")
		return
	}
	for _, fp := range f.basepath {
		fi,err:=os.Stat(fp)
		if err!=nil{
			continue
		}
		if !fi.IsDir(){
			continue
		}
		f.watchPath(fp)
	}
	f.watchDel()
}

//删除判断
func (f *PollWatch) watchDel(){
	if !ClientConfigObj.AllowDelFile{
		return
	}
	if f.pollCount==0{
		return
	}
	LeveldbDB.lock.Lock()
	defer LeveldbDB.lock.Unlock()
	//遍历所有的目录，然后进行判断
	iter := LeveldbDB.db.NewIterator(util.BytesPrefix([]byte("WC_")), nil)
	tempPollCount:=int64(0)
	for iter.Next() {
		bytesBufferRead := bytes.NewBuffer(iter.Value())
		binary.Read(bytesBufferRead, binary.BigEndian, &f.tempValue)
		binary.Read(bytesBufferRead, binary.BigEndian, &tempPollCount)
		if tempPollCount!=f.pollCount{
			p:=string(iter.Key())[3:]
			LeveldbDB.db.Delete(iter.Key(),nil)
			//在上一个批次里，则删除了
			if tempPollCount==f.prePollCount{
				FileChangeMonitorObj.AddPath("del "+p)
			}
		}
	}
	iter.Release()
	iter.Error()
	f.prePollCount=f.pollCount
}

//轮训目录
func (f *PollWatch) watchPath(pa string){
	//通过Walk来遍历目录下f的所有子目录
	filepath.Walk(pa, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() && !ClientConfigObj.AllowDelDir{
			return nil
		}
		fkey:=[]byte("WC_"+p)
		fvue:=info.Size()+info.ModTime().UnixNano()/1e6
		//存入文件大小，最后修改时间，批次3个信息
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian,fvue)
		binary.Write(bytesBuffer, binary.BigEndian, f.pollCount)
		//首次，只是更新目录
		if f.pollCount==0{
			LeveldbDB.Put(fkey,bytesBuffer.Bytes())
			return nil
		}


		//获取文件
		finfo,err:=LeveldbDB.Get(fkey)

		//空的
		if err!=nil{
			//新增
			FileChangeMonitorObj.AddPath("add "+p)
			LeveldbDB.Put(fkey,bytesBuffer.Bytes())
		}else{
			//判断值是否相等，不相等，则认为是变更了
			bytesBufferRead := bytes.NewBuffer(finfo)

			binary.Read(bytesBufferRead, binary.BigEndian, &f.tempValue)
			//修改
			if f.tempValue!=fvue{
				FileChangeMonitorObj.AddPath("m "+p)
			}
			LeveldbDB.Put(fkey,bytesBuffer.Bytes())
		}

		return nil
	})
}

