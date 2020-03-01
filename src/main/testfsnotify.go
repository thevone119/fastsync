package main

import (
	"comm"
	"github.com/fsnotify/fsnotify"
	"os"
	"strconv"
)

func main() {
	//批量创建10万个目录
	if true {
		for i := 0; i < 100; i++ {
			for j := 0; j < 1000; j++ {
				fp := "/home/ccb/notify/" + strconv.FormatInt(int64(i), 10) + "/" + strconv.FormatInt(int64(j), 10)
				os.MkdirAll(fp, os.ModePerm)
			}
		}
	}

	watch, _ := fsnotify.NewWatcher()
	w := comm.Watch{
		Watch: watch,
	}
	w.WatchDir("/home/ccb/notify")
	select {}

}
