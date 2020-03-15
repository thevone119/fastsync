package server

import (
	"comm"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"zinx/zlog"
)

//服务端配置
type serverConfig struct {
	BasePath     string //文件服务的基础路径
	AllowDelFile     bool   //是否允许删除
	AllowDelDir     bool   //是否允许删除
	ConfFilePath string
	UserName     string //服务器的登录用户名
	PassWord     string //服务器的密码
}

/*
	定义一个全局的对象
*/
var ServerConfigObj *serverConfig

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	ServerConfigObj = &serverConfig{
		BasePath:     "/home/ap/ccb/fastsync",
		ConfFilePath: "conf/server.json",
		UserName:     "admin",
		PassWord:     "admin123",
		AllowDelFile:     false,
		AllowDelDir:     false,
	}

	//从配置文件中加载一些用户配置的参数
	ServerConfigObj.Reload()
	ServerConfigObj.BasePath,_=filepath.Abs(ServerConfigObj.BasePath)
	zlog.Debug("serverConfig load BasePath:", ServerConfigObj.BasePath)
}

//读取用户的配置文件
func (g *serverConfig) Reload() {
	spath:=g.ConfFilePath
	if confFileExists, _ := comm.PathExists(spath); confFileExists != true {
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		spath = filepath.Join(path,g.ConfFilePath)
	}

	if confFileExists, _ := comm.PathExists(spath); confFileExists != true {
		return
	}
	zlog.Debug("server conf",spath)
	data, err := ioutil.ReadFile(spath)
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	err = json.Unmarshal(data, g)

}
