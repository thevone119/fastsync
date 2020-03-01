package comm

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
)

//监听文件变化，这个要测试性能，看10W，100W个目录消耗内存情况。
//目前GO有2个监听包https://github.com/howeyc/fsnotify  https://github.com/fsnotify/fsnotify
//需要做一些完整测试。比如移动目录，好像无法监听？

//监听某个目录的变化
func StartFsNotify(fpath string) {

}

type Watch struct {
	Watch *fsnotify.Watcher
}

//通过字符串判断是否目录，就是最后一个如果有文件后缀就判断是目录
func (w *Watch) IsDir(dir string) bool {
	return strings.Index(dir, ".") > 0
}

//监控目录
func (w *Watch) WatchDir(dir string) {
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
			count++
			err = w.Watch.Add(path)
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
			case ev := <-w.Watch.Events:
				{
					if ev.Op&fsnotify.Create == fsnotify.Create {
						fmt.Println("创建文件 : ", ev.Name)
						//这里获取新创建文件的信息，如果是目录，则加入监控中,
						//1.这里还要考虑子目录的问题
						fi, err := os.Stat(ev.Name)
						if err == nil && fi.IsDir() {
							w.Watch.Add(ev.Name)
							fmt.Println("添加监控 : ", ev.Name)
						}
					}
					if ev.Op&fsnotify.Write == fsnotify.Write {
						fmt.Println("写入文件 : ", ev.Name)
					}
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						fmt.Println("删除文件 : ", ev.Name)
						//如果删除文件是目录，则移除监控
						if w.IsDir(ev.Name) {
							w.Watch.Remove(ev.Name)
							fmt.Println("删除监控 : ", ev.Name)
						}
					}
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						fmt.Println("重命名文件 : ", ev.Name)
						if w.IsDir(ev.Name) {
							w.Watch.Remove(ev.Name)
							fmt.Println("删除监控 : ", ev.Name)
						}
					}
					if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
						fmt.Println("修改权限 : ", ev.Name)
					}
				}
			case err := <-w.Watch.Errors:
				{
					fmt.Println("error : ", err)
					//return;
				}
			}
		}
	}()
}
