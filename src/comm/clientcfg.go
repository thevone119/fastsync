package comm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"zinx/zlog"
)

//客户端配置
type clientConfig struct {
	LocalPath        string //本机监控路径
	//全量推送校验规则
	AllowDelFile     bool   //是否允许删除文件
	AllowDelDir     bool   //是否允许删除目录
	filterFiles []string 	//过滤文件类型
	NotifyMonitor []string	//notify监控的路径
	MaxFileCache int64		//最大的单文件缓存，默认1M
	PollMonitor []string	//轮询监控的路径
	PollTime int64			//轮询时间
	LogMonitor  []string	//日志监控路径
	LogCleanDay  int	//日志清空，天，超过X天的日志自动清空，默认90天
	ConfFilePath string
	RemotePath   []string //推送端路径，多个推送端
	remoteName   []string //推送端路径，多个推送端的名称ip+端口+路径
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
		MaxFileCache:1024000,
		LogCleanDay:90,
	}

	//从配置文件中加载一些用户配置的参数
	ClientConfigObj.reload()
	ClientConfigObj.LocalPath,_=filepath.Abs(ClientConfigObj.LocalPath)
	ClientConfigObj.remoteName=make([]string,len(ClientConfigObj.RemotePath))
	//username|pwd|127.0.0.1:9001/
	for i:=0;i<len(ClientConfigObj.RemotePath);i++{
		names := strings.Split(ClientConfigObj.RemotePath[i], "|")
		ClientConfigObj.remoteName[i]=names[2]
	}
	BASE_PATH = ClientConfigObj.LocalPath
	zlog.Debug("SyncConfig load LocalPath:", ClientConfigObj.LocalPath, "RemotePath len:", len(ClientConfigObj.RemotePath))
}
func (g *clientConfig) GetRemoteName(i int) string{
	if i>len(g.remoteName){
		return "null"
	}
	return g.remoteName[i]
}


//读取用户的配置文件
func (g *clientConfig) reload() {

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

//根据绝对路径，获取相对路径
func (g *clientConfig) GetRelativePath(lp string) (string, error) {
	if strings.Index(lp, g.LocalPath) != 0 {
		return "", errors.New("path err:" + lp + ",LocalPath:" + g.LocalPath)
	}
	p := lp[len(g.LocalPath):]
	return p, nil
}
