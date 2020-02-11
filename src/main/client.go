package main

import (
	"client"
	"comm"
	"fmt"
	"os"
	"time"
	"utils"
)

func main() {
	fmt.Println("start...")
	c := client.NewClientUpManager()
	//开启一个客户端监听处理
	time.Sleep(2 * time.Second)
	currtime := time.Now().UnixNano() / 1e6
	c.SyncPath("D:/code")
	fmt.Println("usertime:", time.Now().UnixNano()/1e6-currtime, c.SecId)

	c.SyncFile("UnityPlayer.dll")
	p, _ := os.Getwd()
	fmt.Println("start...,pwd:", p)

	select {}
}

//客户端主程序，无限循环处理
func Process() {

}

//校验某个文件是否需要上传
func checkfile(c *client.NetWork, fp string, ct byte) {
	fmt.Println("fp:", comm.AppendPath(comm.SyncConfigObj.BasePath, fp))
	md5, err := utils.GetFileMd5(comm.AppendPath(comm.SyncConfigObj.BasePath, fp), ct)
	if err != nil {
		fmt.Println("err")
	} else {
		c.Enqueue(comm.NewCheckFileMsg(fp, md5, ct).GetMsg())
	}
}

func callback(ret byte) {

}
