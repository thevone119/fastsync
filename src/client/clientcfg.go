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
	BasePath     string
	ConfFilePath string
	RemotePath   []string
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
		BasePath:     "/test",
		ConfFilePath: "conf/client.json",
	}

	//从配置文件中加载一些用户配置的参数
	ClientConfigObj.reload()
	fmt.Println("SyncConfig load BasePath:", ClientConfigObj.BasePath, "RemotePath len:", len(ClientConfigObj.RemotePath))
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
	if strings.Index(lp, g.BasePath) != 0 {
		return "", errors.New("path err:" + lp)
	}
	p := lp[len(g.BasePath):]

	if strings.Index(p, "/") == 0 {
		p = p[1:]
	}
	return p, nil
}
