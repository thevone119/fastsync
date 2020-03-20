package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"zinx/ziface"
)

/*

*/
type GlobalObj struct {
	/*
		Server
	*/
	TcpServer ziface.IServer //当前的全局Server对象
	Host      string         //当前服务器主机IP
	TcpPort   int            //当前服务器主机监听端口号
	Name      string         //当前服务器名称

	/*

	*/
	Version          string //当前版本号
	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 //业务工作Worker池的数量
	MaxWorkerTaskLen uint32 //业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 //SendBuffMsg发送消息的缓冲最大长度

	/*
		config file path
	*/
	ConfFilePath string

}

/*
	定义一个全局的对象
*/
var GlobalObject *GlobalObj

//判断一个文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

//读取用户的配置文件
func (g *GlobalObj) Reload() {

	spath:=g.ConfFilePath
	if confFileExists, _ := PathExists(spath); confFileExists != true {
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		spath = filepath.Join(path,g.ConfFilePath)
	}

	if confFileExists, _ := PathExists(spath); confFileExists != true {
		return
	}

	data, err := ioutil.ReadFile(spath)
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	err = json.Unmarshal(data, g)
	if err != nil {
		panic(err)
	}

}

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "AiSite FastSync",
		Version:          "V1.11",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          100,
		MaxPacketSize:    10240, //每个数据包的最大大小，10K
		ConfFilePath:     "conf/server.json",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
	}

	//从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
