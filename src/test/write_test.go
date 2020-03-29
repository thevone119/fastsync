package test

import (
	"fmt"
	"os"
	"testing"
)
//单纯文件写入，100万，4秒
func TestWrite(t *testing.T) {
	name:="d:/test/log.txt"
	fileObj,err := os.OpenFile(name,os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	if err != nil {
		fmt.Println("Failed to open the file",err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()
	for i:=0;i<10000*100;i++{
		fileObj.WriteString("testttt\n")
	}



}