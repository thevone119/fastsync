package comm

import (
	"fmt"
	"testing"
)

func TestFileinfo(t *testing.T) {
	f:=NewFFileInfo("E:/go_work/fastsync/fastsyncddd")
	f.ReLoadBase()
	fmt.Println(f.ModTime,f.IsDir,f.IsExist,f.Size)
}
