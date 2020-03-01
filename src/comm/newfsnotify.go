package comm

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
)

//监听文件变化，这个要测试性能，看10W，100W个目录消耗内存情况。
//目前GO有2个监听包https://github.com/howeyc/fsnotify  https://github.com/fsnotify/fsnotify
//需要做一些完整测试。比如移动目录，好像无法监听？

type NewWatch struct {
	Watch *fsnotify.Watcher
}

//监控目录
func (w *NewWatch) WatchDir(dir string) {
	count := 0
	//通过Walk来遍历目录下的所有子目录
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		//这里判断是否为目录，只需监控目录即可
		//目录下的文件也在监控范围内，不需要我们一个一个加
		if info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			w.WatchDir(path)
			count++
			if err != nil {
				return err
			}
			if count%1000 == 0 {
				fmt.Println("监控2 : ", path, count)
			}
		}
		return nil
	})
	go func() {
		for {
			select {
			case ev := <-w.Watch.Event:
				{
					if ev.IsCreate() {
						fmt.Println("创建文件 : ", ev.Name)
						//这里获取新创建文件的信息，如果是目录，则加入监控中
						fi, err := os.Stat(ev.Name)
						if err == nil && fi.IsDir() {
							w.WatchDir(ev.Name)
							fmt.Println("添加监控 : ", ev.Name)
						}
					}
					if ev.IsModify() {
						fmt.Println("写入文件 : ", ev.Name)
					}
					if ev.IsDelete() {
						fmt.Println("删除文件 : ", ev.Name)
						//如果删除文件是目录，则移除监控
						w.Watch.RemoveWatch(ev.Name)
						fmt.Println("删除监控 : ", ev.Name)
					}
					if ev.IsRename() {
						fmt.Println("重命名文件 : ", ev.Name)
						//如果重命名文件是目录，则移除监控
						//注意这里无法使用os.Stat来判断是否是目录了
						//因为重命名后，go已经无法找到原文件来获取信息了
						//所以这里就简单粗爆的直接remove好了
						w.Watch.RemoveWatch(ev.Name)
					}
					if ev.IsModify() {
						fmt.Println("修改权限 : ", ev.Name)
					}
				}
			case err := <-w.Watch.Error:
				{
					fmt.Println("error : ", err)
					return
				}
			}
		}
	}()
}
