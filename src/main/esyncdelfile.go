package main

import (
	"client"
	"fmt"
	"time"
)

func main() {
	currtime := time.Now().UnixNano() / 1e6
	c := client.NewClientUpManager()
	//c.SyncPath("D:/code")
	c.DeleteFile("e:/project/Server.exe")
	fmt.Println("esyncfile use time:", time.Now().UnixNano()/1e6-currtime, c.SecId)

	select {}

}
