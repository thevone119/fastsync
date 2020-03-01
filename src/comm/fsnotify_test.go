package comm

import (
	"github.com/howeyc/fsnotify"
	"os"
	"strconv"
	"testing"
)

func TestFsnotify(t *testing.T) {
	//批量创建100万个目录
	for i := 0; i < 1000; i++ {
		for j := 0; j < 1000; j++ {
			fp := "E:/notify/" + strconv.FormatInt(int64(i), 10) + "/" + strconv.FormatInt(int64(j), 10)
			os.MkdirAll(fp, os.ModePerm)
		}
	}

	watch, _ := fsnotify.NewWatcher()
	w := NewWatch{
		Watch: watch,
	}
	w.WatchDir("E:/notify")
	select {}
}
