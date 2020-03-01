package client

import (
	"comm"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

//客户端配置
type clientConfig struct {
	LocalPath        string //本机监控路径
	LocalPathMonitor bool   //是否开启本机监控
	//全量推送校验规则
	AllowDel     bool //是否允许删除
	ConfFilePath string
	RemotePath   []string //推送端路径，多个推送端

}

/*
	定义一个全局的对象
*/
var ClientConfigObj *clientConfig

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	ClientConfigObj = &clientConfig{
		LocalPath:    "/test",
		ConfFilePath: "conf/client.json",
	}

	//从配置文件中加载一些用户配置的参数
	ClientConfigObj.reload()
	fmt.Println("SyncConfig load LocalPath:", ClientConfigObj.LocalPath, "RemotePath len:", len(ClientConfigObj.RemotePath))
}

//读取用户的配置文件
func (g *clientConfig) reload() {

	if confFileExists, _ := comm.PathExists(g.ConfFilePath); confFileExists != true {
		//fmt.Println("Config File ", g.ConfFilePath , " is not exist!!")
		return
	}

	data, err := ioutil.ReadFile(g.ConfFilePath)
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	err = json.Unmarshal(data, g)
}

//根据绝对路径，获取相对路径
func (g *clientConfig) GetRelativePath(lp string) (string, error) {
	if strings.Index(lp, g.LocalPath) != 0 {
		return "", errors.New("path err:" + lp + ",LocalPath:" + g.LocalPath)
	}
	p := lp[len(g.LocalPath):]

	if strings.Index(p, "/") == 0 {
		p = p[1:]
	}
	return p, nil
}
