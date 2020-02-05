package main

import "client"

func main(){
	//开启一个客户端监听处理
	c:=client.NewNetWork("127.0.0.1",8999)
	go c.Process()


	select {
	}
}

//客户端主程序，无限循环处理
func Process(){

}

//校验某个文件是否需要上传
func checkfile(fp string){
	
}