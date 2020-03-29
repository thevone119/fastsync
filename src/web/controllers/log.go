package controllers

//日志查询
type LogController struct {
	BaseController
}

func (self *LogController) List() {
	self.Data["pageTitle"] = "传输日志"
	self.display()
}

func (self *LogController) Start() {
	self.Data["pageTitle"] = "控制面板"
	self.display()
}