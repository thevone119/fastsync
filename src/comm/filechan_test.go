package comm

import (
	"fmt"
	"testing"
)

func TestFilechan(t *testing.T) {
	f := NewFileChan("E:/TEST/test2", "*.txt")
	FileDB.Open()
	for {
		select {
		case l := <-f.LineChan:
			fmt.Println(l)
			FileDB.Commit()
		}
	}

}
