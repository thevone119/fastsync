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
	c.MoveFile("e:/project/发布20200228", "e:/project/发布202002282")
	fmt.Println("esyncfile use time:", time.Now().UnixNano()/1e6-currtime, c.SecId)

	select {}

}
