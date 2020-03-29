package test

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

func main() {
	go gocopy(1)
	go gocopy(2)
	go gocopy(3)
	time.Sleep(time.Second*3)
}

func gocopy(i int){
	src, _ := os.Open("/home/"+strconv.FormatInt(int64(i),10))
	defer src.Close()
	dst, err := os.OpenFile("/home/10", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err!=nil{
		fmt.Println("err",err)
	}
	defer dst.Close()
	fmt.Println("start",time.Now().UnixNano())
	r,err:=io.Copy(dst, src)
	fmt.Println(r,err)
	fmt.Println("end",time.Now().UnixNano())
}