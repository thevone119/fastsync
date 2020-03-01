package comm

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestTempMap(t *testing.T) {
	go goWrite()
	go goWrite()
	go goWrite()
	go goRead()
	go goRead()
	go goRead()

	select {}
}

func goWrite() {
	rand.Seed(time.Now().UnixNano())
	for {
		//随机放入100万个字符串到tempmap,每6秒放一次，测试内存
		for i := 0; i < 1000000; i++ {
			TempMap.Put("svn-base_"+strconv.FormatInt(rand.Int63(), 10)+"_"+string(i), "v", 2)
		}
		fmt.Println("jianhang/eportalapp/.svn/pristine/48/489ed6a9c610bf03a4b5b398592c43c382e54be0.svn-base_" + strconv.FormatInt(rand.Int63(), 10))
		fmt.Println("put end...", TempMap.len())
		time.Sleep(time.Second)
	}
}

//同时开启读取操作，每2秒循环1万次的读取
func goRead() {
	rand.Seed(time.Now().UnixNano())
	for {
		for i := 0; i < 10000; i++ {
			TempMap.Get("svn-base_" + string(rand.Int63()))
		}
		fmt.Println("get end...")
		time.Sleep(time.Second)
	}
}
