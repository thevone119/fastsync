package comm

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"time"
	"zinx/zlog"
)

//用bolt做的简单的key/value存储

//全局变量，直接使用
var FileDB *filedb

type filedb struct {
	path   string
	db     *bolt.DB
	isopen bool
}

/*
	提供init方法，默认加载
*/
func init() {
	//初始化全局变量，设置一些默认值
	FileDB = &filedb{
		path: "fastsync.db",
	}
	FileDB.open()
}

//只开，不关，只有一个应用使用，引用退出，自动就退出了
func (f *filedb) open() {
	if f.isopen {
		return
	}
	db, err := bolt.Open(f.path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		zlog.Error("open db err", err)
		return
	}
	f.db = db
	f.isopen = true
}

func (f *filedb) PutString(bucket string, key string, value string) error {
	err := f.db.Update(func(tx *bolt.Tx) error {
		// 创建一个桶
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
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
	})
	return err
}

func (f *filedb) PutInt64(bucket string, key string, value int64) error {
	err := f.db.Update(func(tx *bolt.Tx) error {
		// 创建一个桶
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			zlog.Error("filedb PutString err", err)
			return err
		}
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, value)
		err = b.Put([]byte(key), bytesBuffer.Bytes())
		if err != nil {
			zlog.Error("filedb PutString2 err", err)
			return err
		}
		return nil
	})
	return err
}

func (f *filedb) GetString(bucket string, key string) (string, error) {
	var ret = ""
	err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		ret = string(v[:])
		return nil
	})
	return ret, err
}

func (f *filedb) GetInt64(bucket string, key string) (int64, error) {
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

func (f *filedb) RemoveString(bucket string, key string) error {
	err := f.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		b.Delete([]byte(key))
		return nil
	})
	return err
}
