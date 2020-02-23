package main

import (
	"client"
	"fmt"
	"time"
	"zinx/zlog"
)

func main() {

	goSyncPath()
	select {}

}

func goSyncPath() {
	zlog.CloseDebug()
	c := client.NewClientUpManager()
	fmt.Println("start:")

	for {
		currtime := time.Now().UnixNano() / 1e6
		//c.SyncPath("D:/code")
		c.SyncPath("e:/project/")
		fmt.Println("esyncfile use time:", time.Now().UnixNano()/1e6-currtime, c.SecId)
		time.Sleep(time.Second * 60 * 5)
	}

}
