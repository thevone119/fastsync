package comm

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSyncConfig(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	fmt.Println(rand.Int63(), rand.Int63())
	fmt.Println(rand.Int63(), rand.Int63())
	var s = "d:/test/ttttt/ttttt/1.txt"
	_, err := os.Stat(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(s[:strings.LastIndex(s, "/")])

	by := make([]byte, 10)
	by2 := by[0:10]
	fmt.Println("blen:", len(by2))

	os.MkdirAll("d:/test/ttttt/ttttt/", os.ModePerm)
	if os.IsNotExist(err) {

	}
	fmt.Println(AppendPath("d:/test/ttttt/ttttt/", "/test/ttttt/ttttt/1.txt"))

	vs := "fasljfa/"
	fmt.Println("tt:", vs[:len(vs)-1])

	dstwrite, err := os.Create(s)
	if err != nil {
		fmt.Println(err)
	}
	dstwrite.WriteString("tetst")
	defer dstwrite.Close()

}
