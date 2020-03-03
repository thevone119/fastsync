package test

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"testing"
	"zinx/zlog"
)

func TestChan(t *testing.T) {
	ch := make(chan int, 10)
	ch <- 1
	ch <- 1
	ch <- 1
	ch <- 1
	ch <- 1

	for i := 0; i < 0; i++ {
		<-ch
	}
	fmt.Println("ch len:", len(ch))
	n, err := doUploadChan()
	fmt.Println("ch len:", n, err)
	testpath()
}

func testpath() {

	fmt.Println("path:", path.Ext("d:/sal.dfj/daf.txt"))
	s, _ := filepath.EvalSymlinks("D:/video///[BT200.COM][新金瓶梅3D][720P高清完整版][3.49GB]/d.tt")
	fmt.Println("path:", s)
}

func doUploadChan() (retb byte, err error) {

	defer func() {
		//恢复程序的控制权
		if p := recover(); p != nil {
			str, ok := p.(string)
			if ok {
				err = errors.New(str)
			} else {
				err = errors.New("panic")
			}
			retb = 100
			zlog.Error("文件上传发送意外错误", err)
		}
	}()

	return doUploadChan2()
}

func doUploadChan2() (retb byte, err error) {
	panic("异常错误")
	return 0, nil
}
