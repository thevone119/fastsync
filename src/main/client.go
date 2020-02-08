package main

import (
	"client"
	"comm"
	"fmt"
	"time"
	"utils"
)

func main(){
	//开启一个客户端监听处理
	c:=client.NewNetWork("127.0.0.1",8999)
	go c.Process()
	time.Sleep(2 * time.Second)
	//测试10万个同步请求的耗时(28秒)
	t1:=time.Now().Unix()
	for i:=0;i<100000;i++{
		c.Request(comm.NewKeepAliveMsg(1).GetMsg())
	}
	fmt.Println("usertime:",time.Now().Unix()-t1)

	checkfile(c,"UnityPlayer.dll",2)
	fload:=client.NewFileUpload(c,20,"")
	go fload.Upload("D:/test/UnityPlayer.dll","UnityPlayer.dll",1,callback)
	//fload.Upload("D:/test/UnityPlayer.dll","UnityPlayer.dll",1,callback)
	//fload.Upload("D:/test/UnityPlayer.dll","UnityPlayer.dll",1,callback)
	//fload.Upload("D:/test/UnityPlayer.dll","UnityPlayer.dll",1,callback)
	select {
	}
}

//客户端主程序，无限循环处理
func Process(){

}

//校验某个文件是否需要上传
func checkfile(c *client.NetWork,fp string,ct byte){
	fmt.Println("fp:",comm.AppendPath(comm.SyncConfigObj.BasePath,fp))
	md5,err:=utils.GetFileMd5(comm.AppendPath(comm.SyncConfigObj.BasePath,fp),ct)
	if err !=nil{
		fmt.Println("err")
	}else{
		c.Enqueue(comm.NewCheckFileMsg(fp,md5,ct).GetMsg())
	}
}

func callback(ret byte){

}