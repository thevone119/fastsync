package comm

import (
	"encoding/json"
	"io/ioutil"
)

type SyncConfig struct {
	BasePath string
	ConfFilePath string
}

/*
	定义一个全局的对象
*/
var SyncConfigObj *SyncConfig


/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	SyncConfigObj = &SyncConfig{
		BasePath:             "ZinxServerApp",
		ConfFilePath:"conf/sync.json",
	}

	//从配置文件中加载一些用户配置的参数
	SyncConfigObj.Reload()
}

//读取用户的配置文件
func (g *SyncConfig) Reload() {

	if confFileExists, _ := PathExists(g.ConfFilePath); confFileExists != true {
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


