package comm

import (
	"encoding/json"
	"io/ioutil"
	"zinx/zlog"
)

//通用配置，公共组件配置，比如log日志配置
type CommConfig struct {
	ConfFilePath string
	/*
		logger
	*/
	LogDir        string //日志所在文件夹 默认"./log"
	LogFile       string //日志文件名称   默认""  --如果没有设置日志文件，打印信息将打印至stderr
	LogDebugClose bool   //是否关闭Debug日志级别调试信息 默认false  -- 默认打开debug信息
}


/*
	定义一个全局的对象
*/
var CommConfigObj *CommConfig


//读取用户的配置文件
func (g *CommConfig) Reload() {

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
	if err != nil {
		panic(err)
	}

	//Logger 设置
	if g.LogFile != "" {
		zlog.SetLogFile(g.LogDir, g.LogFile)
	}
	if g.LogDebugClose == true {
		zlog.CloseDebug()
	}
}

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	CommConfigObj = &CommConfig{
		ConfFilePath:     "conf/comm.json",
		LogDir:           "./syslog",
		LogFile:          "",
		LogDebugClose:    false,
	}

	//从配置文件中加载一些用户配置的参数
	CommConfigObj.Reload()
}
