package comm

import (
	"testing"
)

func TestFilechan(t *testing.T) {
	LeveldbDB.Open()
	f := NewFileChan("E:/TEST/test2", "*.txt")
	f.Start()


}
