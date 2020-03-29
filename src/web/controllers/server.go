package controllers

import (
	"client"
	"strings"
)

//服务器状态查询
type ServerController struct {
	BaseController
}

func (self *ServerController) List() {
	self.Data["pageTitle"] = "服务器状态"

	self.display()
}

//查询服务器列表
func (self *ServerController) Table() {
	//列表
	/*page, err := self.GetInt("page")
	if err != nil {
		page = 1
	}*/
	limit, err := self.GetInt("limit")
	if err != nil {
		limit = 30
	}

	realName := strings.TrimSpace(self.GetString("realName"))

	StatusText := make(map[int]string)
	StatusText[0] = "<font color='red'>禁用</font>"
	StatusText[1] = "正常"

	self.pageSize = limit
	//查询条件
	filters := make([]interface{}, 0)
	//
	if realName != "" {
		filters = append(filters, "real_name__icontains", realName)
	}
	//开启一个客户端监听处理
	if client.ClientObj==nil{
		return
	}

	list := make([]map[string]interface{}, len(client.ClientObj.Client.RemoteUpLoad))
	for k, v := range client.ClientObj.Client.RemoteUpLoad {
		row := make(map[string]interface{})
		row["Id"] = v.Netclient.Id
		row["Name"] = v.Name
		row["SendSpeed"] = v.SendSpeed
		row["MinSendSpeed"] = v.MinSendSpeed
		row["MaxSendSpeed"] = v.MaxSendSpeed
		row["SuccSendCount"] = v.SuccSendCount
		row["ErrorSendCount"] = v.ErrorSendCount
		row["Connected"] = v.Netclient.Connected
		row["PingTime"] = v.Netclient.PingTime
		row["MaxPingTime"] = v.Netclient.MaxPingTime
		row["ConnCount"] = v.Netclient.ConnCount
		row["SuccConnCount"] = v.Netclient.SuccConnCount
		row["OnlineTime"] = v.Netclient.OnlineTime
		row["OfflineTime"] = v.Netclient.OfflineTime
		list[k] = row
	}
	self.ajaxList("成功", MSG_OK, int64(len(list)), list)
}

func (self *ServerController) Start() {
	self.Data["pageTitle"] = "控制面板"
	self.display()
}