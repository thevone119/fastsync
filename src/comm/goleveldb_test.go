package comm

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strconv"
	"testing"
	"time"
	"zinx/zlog"
)
//测试结论，goleveldb性能非常均衡的一个数据库，内存占用，磁盘占用都比较低
//bolt速度更快，100万随机读取不到1秒完成。但是写入需要提交后才能读取。基于MMAP，如果每次都写入，则写入速度非常慢。适用于一次写入，频繁读取的场景
func TestLeveldb(t *testing.T) {
	LeveldbDB.DBPath="d:\\test\\test2\\db";
	LeveldbDB.Open()

	go testwrite()
	go testread()
	select {

	}
}


//写入性能测试，写入100万数据,100万7秒
func testwrite(){
	// 创建或打开一个数据库
	currtime:=time.Now()
	for i:=0;i<10000*100;i++{
		LeveldbDB.Put([]byte("asfdasdfasdfasdfwefasdfdghrtyrsfagadarwebgbferevtretetetevercecrqwrcwqcrqwrxdfvdkey_"+strconv.FormatInt(int64(i),10)), []byte("value_"+strconv.FormatInt(int64(i),10)))
	}
	zlog.Info("wirte end user time",time.Now().Sub(currtime))
}

//读取性能测试，读取100万数据,随机读，4秒
func testread(){
	// 创建或打开一个数据库
	currtime:=time.Now()
	for i:=0;i<10000*100;i++{
		_, err := LeveldbDB.Get([]byte("asfdasdfasdfasdfwefasdfdghrtyrsfagadarwebgbferevtretetetevercecrqwrcwqcrqwrxdfvdkey_"+strconv.FormatInt(int64(i),10)))
		if err!=nil{
			fmt.Println(err)
		}
	}
	zlog.Info("wirte end user time",time.Now().Sub(currtime))
}

//简单的全操作测试
func test1(){
	fmt.Println("hello")

	// 创建或打开一个数据库
	db, err := leveldb.OpenFile("C:/Users/Administrator/AppData/Local/Temp/___TestLeveldb_in_comm.exe.leveldb", nil)
	if err != nil {
		panic(err)
	}

	//defer db.Close()

	// 存入数据
	db.Put([]byte("1"), []byte("6"), nil)
	db.Put([]byte("2"), []byte("7"), nil)
	db.Put([]byte("3"), []byte("8"), nil)
	db.Put([]byte("foo-4"), []byte("9"), nil)
	db.Put([]byte("5"), []byte("10"), nil)
	db.Put([]byte("6"), []byte("11"), nil)
	db.Put([]byte("moo-7"), []byte("12"), nil)
	db.Put([]byte("8"), []byte("13"), nil)

	// 遍历数据库内容
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Printf("[%s]:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		panic(err)

	}

	fmt.Println("***************************************************")

	// 删除某条数据
	err = db.Delete([]byte("2"), nil)

	// 读取某条数据
	data, err := db.Get([]byte("2"), nil)
	fmt.Printf("[2]:%s:%s\n", data, err)

	// 根据前缀遍历数据库内容
	fmt.Println("***************************************************")
	iter = db.NewIterator(util.BytesPrefix([]byte("foo-")), nil)
	for iter.Next() {
		fmt.Printf("[%s]:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	err = iter.Error()

	// 遍历从指定 key
	fmt.Println("***************************************************")
	iter = db.NewIterator(nil, nil)
	for ok := iter.Seek([]byte("5")); ok; ok = iter.Next() {
		fmt.Printf("[%s]:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	err = iter.Error()

	// 遍历子集范围
	fmt.Println("***************************************************")
	iter = db.NewIterator(&util.Range{Start: []byte("foo"), Limit: []byte("loo")}, nil)
	for iter.Next() {
		fmt.Printf("[%s]:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	err = iter.Error()

	// 批量操作
	fmt.Println("***************************************************")
	batch := new(leveldb.Batch)
	batch.Put([]byte("foo"), []byte("value"))
	batch.Put([]byte("bar"), []byte("another value"))
	batch.Delete([]byte("baz"))
	err = db.Write(batch, nil)

	// 遍历数据库内容
	iter = db.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Printf("[%s]:%s\n", iter.Key(), iter.Value())
	}
	iter.Release()
	err = iter.Error()
}