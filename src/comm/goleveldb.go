package comm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"sync"
	"zinx/zlog"
)

//全局变量，直接使用
var LeveldbDB *levelDB


type levelDB struct {
	DBPath string
	db *leveldb.DB
	isopen bool
	lock        sync.RWMutex
}


/*
	提供init方法，默认加载
*/
func init() {
	//初始化全局变量，设置一些默认值
	//每个进程一个数据库文件。不能开多个进程，否则就会发生数据文件锁哦，这个数据文件无法多个进程共享的
	LeveldbDB = &levelDB{
		DBPath: "leveldb",
	}
	path, err := os.Executable()
	if err != nil {
		zlog.Error("level db get path err",err)
	}else {
		LeveldbDB.DBPath = path+".leveldb"
	}
}

//打开
func (d *levelDB) Open(){
	if d.isopen{
		return
	}
	db, err := leveldb.OpenFile(d.DBPath, nil)
	if err!=nil{
		zlog.Error("level db open err",err)
	}else{
		zlog.Info("db open:",d.DBPath)
		d.db=db
		d.isopen=true
	}
}

//关闭
func (d *levelDB) Close(){
	if !d.isopen{
		return
	}
	err:=d.db.Close()
	if err!=nil{
		zlog.Error("level db close err",err)
	}else{
		d.isopen=false
	}
}

//一些复杂操作，直接取出DB进行操作
func (d *levelDB) GetDB()  *leveldb.DB{
	return d.db
}

//放入
func (d *levelDB) Put(k []byte ,v []byte) error{
	if !d.isopen{
		return errors.New("db is not open")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.db.Put(k, v, nil)
}
//取出
func (d *levelDB) Get(k []byte)(value []byte, err error){
	if !d.isopen{
		return []byte(""),errors.New("db is not open")
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.db.Get(k, nil)
}
//删除
func (d *levelDB) Del(k []byte) error{
	if !d.isopen{
		return errors.New("db is not open")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.db.Delete(k,nil)
}

//放入INT64
func (d *levelDB) PutInt64(k []byte ,v int64) error{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, v)
	return d.Put(k,bytesBuffer.Bytes())
}

//取出INT64
func (d *levelDB) GetInt64(k []byte)(value int64, err error){
	v,err:=d.Get(k)
	bytesBuffer := bytes.NewBuffer(v)
	binary.Read(bytesBuffer, binary.BigEndian, &value)
	return
}



