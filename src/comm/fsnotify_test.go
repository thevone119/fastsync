package comm

import (
	"testing"
)

func TestFsnotify(t *testing.T) {
	//批量创建100万个目录
	fs := NewFSWatch("d:/video")
	fs.Start()

	select {}
}
