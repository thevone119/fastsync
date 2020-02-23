package client

import (
	"fmt"
	"testing"
	"utils"
)

func TestLocalfile(t *testing.T) {
	lp := "E:/project2/jianhang/eportal/.svn/pristine/6e/6e3cdceb03bd631cf10aed988464b81d6c7d5d1a.svn-base"
	ul := NewLocalFile(lp, lp, 2)
	fmt.Println(ul.FileMd5)
	fmt.Println(utils.GetFileMd5(lp, 2))
	//time.Sleep(time.Second)
}
