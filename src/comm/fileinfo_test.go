package comm

import (
	"fmt"
	"testing"
)

func TestFileinfo(t *testing.T) {
	p:="e:/project/test2/123/test.txt"
	f:=NewFFileInfo(p)
	f.ReLoadBase()
	fmt.Println(p,f.IsExist)

}
