package diskqueue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

//这个文件MQ,文件大小设置为1M-2M，同步数设置为1-2万，时间设置为2秒，比较合适
//这个插入的性能也太差了吧。100万15秒
//出队列，100万，4秒
func TestDiskQueue2(t *testing.T) {
	l := NewTestLogger(t)

	dqName := "test_disk_queue"
	tmpDir:="d:/test/mq"
	fmt.Println(tmpDir)

	fmt.Println(rand.Int())
	fmt.Println(rand.Int())
	//defer os.RemoveAll(tmpDir)
	//1M
	dq := New(dqName, tmpDir, 1024*1024, 4, 1<<10, 20000, 2*time.Second, l)
	defer dq.Close()

	count:=0
	for {
		if dq.Depth()<=0{
			break
		}
		<-dq.ReadChan()
		count++
	}

	fmt.Println(count)


}
