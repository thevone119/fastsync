package test

import (
	"io"
	"os"
	"testing"
	"time"
)

func TestLog(t *testing.T) {

	f,_ := os.Create("e:/project/test2/testlog.out") //打开文件

	for{
		io.WriteString(f, "e:/project/test2/123/test.txt\n")
		time.Sleep(time.Millisecond*500)
	}


}