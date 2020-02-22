package main

import (
	"client"
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("start...")
	c := client.NewClientUpManager()
	//开启一个客户端监听处理
	time.Sleep(2 * time.Second)
	currtime := time.Now().UnixNano() / 1e6
	//c.SyncPath("D:/code")
	fmt.Println("usertime:", time.Now().UnixNano()/1e6-currtime, c.SecId)

	//c.SyncFile("UnityPlayer.dll")
	p, _ := os.Getwd()
	fmt.Println("start...,pwd:", p)

	select {}
}
