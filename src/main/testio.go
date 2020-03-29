package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

var sg=sync.WaitGroup{}
func main() {

	sg.Add(3)
	go gocopy(1)
	time.Sleep(time.Second)
	go gocopy(2)
	time.Sleep(time.Second)
	go gocopy(3)

	sg.Wait()
}

func gocopy(i int){
	dst, err := os.OpenFile("/home/10", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err!=nil{
		fmt.Println("err",err)
	}
	defer dst.Close()
	for j:=0;j<10;j++{
		r,err:=dst.WriteString(strconv.FormatInt(int64(j),10)+"test_"+strconv.FormatInt(int64(i),10)+"\n")
		fmt.Println(r,err)
		time.Sleep(time.Second)
	}
	sg.Done()
}