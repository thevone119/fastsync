package comm

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSyncConfig(t *testing.T) {

	var s = "d:/test/ttttt/ttttt/1.txt"
	_, err := os.Stat(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(s[:strings.LastIndex(s, "/")])

	os.MkdirAll("d:/test/ttttt/ttttt/", os.ModePerm)
	if os.IsNotExist(err) {

	}
	fmt.Println(AppendPath("d:/test/ttttt/ttttt/", "/test/ttttt/ttttt/1.txt"))

	dstwrite, err := os.Create(s)
	if err != nil {
		fmt.Println(err)
	}
	dstwrite.WriteString("tetst")
	defer dstwrite.Close()
}
