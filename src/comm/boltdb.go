package comm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/boltdb/bolt"
	"os"
	"sync"
	"time"
	"zinx/zlog"
)

//用bolt做的简单的key/value存储

//全局变量，直接使用
var BoltDB *boltDB

type boltDB struct {
	DBPath string
	db     *bolt.DB
	isopen bool
	lastWriteTime time.Time
	tx     *bolt.Tx
	istx bool
	lock        sync.RWMutex
}

/*
	提供init方法，默认加载
*/
func init() {
	//初始化全局变量，设置一些默认值
	//每个进程一个数据库文件。不能开多个进程，否则就会发生数据文件锁哦，这个数据文件无法多个进程共享的
	BoltDB = &boltDB{
		DBPath: "fastsync.db",
		lastWriteTime:time.Now(),
	}
	path, err := os.Executable()
	if err != nil {
		zlog.Error("open db error")
		return
	}
	BoltDB.DBPath = path + ".boltdb"
}

//只开，不关，只有一个应用使用，应用退出，自动就退出关闭了
func (f *boltDB) Open() {
	if f.isopen {
		return
	}
	db, err := bolt.Open(f.DBPath, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		zlog.Error("open db err", err)
		return
	}
	zlog.Info(BoltDB.DBPath, " open")
	f.db = db
	f.isopen = true
}
func (f *boltDB) Close(){
	if !f.isopen {
		return
	}
	if f.istx{
		f.tx.Commit()
		f.istx=false
	}
	f.db.Close()
	f.isopen=false
}

//创建数据库桶
func (f *boltDB) CreateBucketIfNotExists(buc string){
	// Create a bucket using a read-write transaction.
	if err :=  f.db.Update(func(tx *bolt.Tx) error {
	_, err := tx.CreateBucketIfNotExists([]byte(buc))
	return err
	});
	err != nil {
		//log.Fatal(err)
		zlog.Error(err)
	}
}

func (f *boltDB) GetDB()  *bolt.DB{
	return f.db
}

func (f *boltDB) Begin(){
	if !f.isopen{
		return
	}
	if !f.istx{
		f.tx,_ = f.db.Begin(true)
		f.istx = true
		f.lastWriteTime = time.Now()
	}
}

//3秒最多提交一次
func (f *boltDB) Commit(){
	if !f.isopen{
		return
	}
	if f.istx && time.Now().Sub(f.lastWriteTime)>time.Second*3{
		f.tx.Commit()
		f.istx=false
	}
}
func (f *boltDB) ForceCommit(){
	if f.istx{
		f.tx.Commit()
		f.istx=false
	}
}





func (f *boltDB) PutString(bucket string, key string, value string) error {
	if !f.isopen{
		zlog.Error("filedb not open db")
		return errors.New("not open db")
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Begin()
	defer f.Commit()
	// 创建一个桶
	b, err := f.tx.CreateBucketIfNotExists([]byte(bucket))
	if err != nil {
		zlog.Error("filedb PutString err", err)
		return err
	}
	err = b.Put([]byte(key), []byte(value))
	if err != nil {
		zlog.Error("filedb PutString2 err", err)
		return err
	}
	return nil
}

func (f *boltDB) PutInt64(bucket string, key string, value int64) error {
	if !f.isopen{
		return errors.New("not open db")
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Begin()
	defer f.Commit()

	// 创建一个桶
	b, err := f.tx.CreateBucketIfNotExists([]byte(bucket))
	if err != nil {
		zlog.Error("filedb PutInt64 err", err)
		return err
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, value)

	err = b.Put([]byte(key), bytesBuffer.Bytes())
	if err != nil {
		zlog.Error("filedb PutInt642 err", err)
		return err
	}
	return err
}

func (f *boltDB) GetString(bucket string, key string) (string, error) {
	if !f.isopen{
		return "",errors.New("not open db")
	}
	f.lock.RLock()
	defer f.lock.RUnlock()
	var ret = ""
	err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		ret = string(v)
		return nil
	})
	return ret, err
}

func (f *boltDB) GetInt64(bucket string, key string) (int64, error) {
	if !f.isopen{
		return int64(-1),errors.New("not open db")
	}
	f.lock.RLock()
	defer f.lock.RUnlock()
	var ret = int64(-1)
	err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		bytesBuffer := bytes.NewBuffer(v)
		var x int64
		binary.Read(bytesBuffer, binary.BigEndian, &x)
		ret = x
		return nil
	})
	return ret, err
}

func (f *boltDB) RemoveString(bucket string, key string) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Begin()
	defer f.Commit()

	b := f.tx.Bucket([]byte(bucket))
	if b == nil {
		return nil
	}
	return b.Delete([]byte(key))
}
