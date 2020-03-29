package main

import (
	"comm"
	"github.com/howeyc/fsnotify"
	"os"
	"strconv"
)

func main() {
	//批量创建100万个目录
	if false {
		for i := 0; i < 1000; i++ {
			for j := 0; j < 1000; j++ {
				fp := "/home/ccb/notify/" + strconv.FormatInt(int64(i), 10) + "/" + strconv.FormatInt(int64(j), 10)
				os.MkdirAll(fp, os.ModePerm)
			}
		}
	}

	watch, _ := fsnotify.NewWatcher()
	w := comm.NewWatch{
		Watch: watch,
	}
	w.WatchDir("/home/ccb/notify")
	select {}

}
